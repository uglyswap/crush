package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SkillSource represents a source location for skills.
type SkillSource string

const (
	SourceLocal   SkillSource = "local"   // Built-in skills
	SourceUser    SkillSource = "user"    // User's ~/.claude/skills/
	SourceProject SkillSource = "project" // Project's .claude/skills/
)

// SkillLoader loads skills from various sources.
type SkillLoader struct {
	sources  map[SkillSource]string
	registry *SkillRegistry
}

// NewSkillLoader creates a new skill loader.
func NewSkillLoader(registry *SkillRegistry) *SkillLoader {
	return &SkillLoader{
		sources:  make(map[SkillSource]string),
		registry: registry,
	}
}

// AddSource adds a skill source directory.
func (l *SkillLoader) AddSource(source SkillSource, path string) {
	l.sources[source] = path
}

// ConfigureDefaults sets up default skill directories.
func (l *SkillLoader) ConfigureDefaults(projectRoot string) error {
	// User skills directory
	homeDir, err := os.UserHomeDir()
	if err == nil {
		userSkillsDir := filepath.Join(homeDir, ".claude", "skills")
		l.sources[SourceUser] = userSkillsDir
	}

	// Project skills directory
	if projectRoot != "" {
		projectSkillsDir := filepath.Join(projectRoot, ".claude", "skills")
		l.sources[SourceProject] = projectSkillsDir
	}

	return nil
}

// LoadAll loads all skills from all configured sources.
func (l *SkillLoader) LoadAll() error {
	var errors []string

	for source, path := range l.sources {
		if err := l.loadFromDirectory(source, path); err != nil {
			errors = append(errors, fmt.Sprintf("%s (%s): %v", source, path, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors loading skills: %s", strings.Join(errors, "; "))
	}

	return nil
}

// loadFromDirectory loads all skills from a directory.
func (l *SkillLoader) loadFromDirectory(source SkillSource, dir string) error {
	// Check if directory exists
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Directory doesn't exist, skip
		}
		return fmt.Errorf("failed to stat directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", dir)
	}

	// Walk the directory
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process markdown files
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".md" && ext != ".markdown" {
			return nil
		}

		// Load skill
		skill, err := l.loadSkillFile(path, source)
		if err != nil {
			// Log error but continue loading other skills
			fmt.Printf("Warning: failed to load skill from %s: %v\n", path, err)
			return nil
		}

		// Register skill
		l.registry.Register(skill)

		return nil
	})
}

// loadSkillFile loads a skill from a file.
func (l *SkillLoader) loadSkillFile(path string, source SkillSource) (*Skill, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	skill, err := ParseSkill(string(content), string(source), path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse skill: %w", err)
	}

	// Get file modification time
	info, err := os.Stat(path)
	if err == nil {
		skill.LastModified = info.ModTime()
	}

	return skill, nil
}

// LoadSkillByName loads a specific skill by name from any source.
func (l *SkillLoader) LoadSkillByName(name string) (*Skill, error) {
	// First check registry
	if skill := l.registry.Get(name); skill != nil {
		return skill, nil
	}

	// Try to find in sources (project > user > local priority)
	sourcePriority := []SkillSource{SourceProject, SourceUser, SourceLocal}

	for _, source := range sourcePriority {
		dir, ok := l.sources[source]
		if !ok {
			continue
		}

		// Try different file patterns
		patterns := []string{
			filepath.Join(dir, name+".md"),
			filepath.Join(dir, name, "skill.md"),
			filepath.Join(dir, name, "index.md"),
			filepath.Join(dir, name+".markdown"),
		}

		for _, pattern := range patterns {
			if _, err := os.Stat(pattern); err == nil {
				skill, err := l.loadSkillFile(pattern, source)
				if err != nil {
					return nil, err
				}
				l.registry.Register(skill)
				return skill, nil
			}
		}
	}

	return nil, fmt.Errorf("skill not found: %s", name)
}

// ReloadSkill reloads a skill from its source.
func (l *SkillLoader) ReloadSkill(name string) (*Skill, error) {
	existing := l.registry.Get(name)
	if existing == nil {
		return l.LoadSkillByName(name)
	}

	// Reload from the same path
	skill, err := l.loadSkillFile(existing.Path, SkillSource(existing.Source))
	if err != nil {
		return nil, fmt.Errorf("failed to reload skill: %w", err)
	}

	l.registry.Register(skill)
	return skill, nil
}
