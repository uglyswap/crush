// Package skills provides a system for loading and invoking specialized capabilities.
package skills

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Skill represents a specialized capability that can be invoked.
type Skill struct {
	Name          string   `json:"name" yaml:"name"`
	Description   string   `json:"description" yaml:"description"`
	AllowedTools  []string `json:"allowed_tools,omitempty" yaml:"allowed-tools,omitempty"`
	Content       string   `json:"content" yaml:"content"`
	Source        string   `json:"source,omitempty"` // local, user, project
	Path          string   `json:"path,omitempty"`
	LastModified  time.Time `json:"last_modified,omitempty"`
}

// SkillMetadata contains metadata extracted from skill frontmatter.
type SkillMetadata struct {
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	AllowedTools []string `yaml:"allowed-tools"`
}

// ParseSkill parses a skill from markdown content with YAML frontmatter.
func ParseSkill(content, source, path string) (*Skill, error) {
	// Extract frontmatter
	frontmatter, body, err := extractFrontmatter(content)
	if err != nil {
		return nil, err
	}

	// Parse frontmatter
	meta, err := parseFrontmatter(frontmatter)
	if err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	if meta.Name == "" {
		return nil, fmt.Errorf("skill name is required")
	}

	return &Skill{
		Name:         meta.Name,
		Description:  meta.Description,
		AllowedTools: meta.AllowedTools,
		Content:      body,
		Source:       source,
		Path:         path,
		LastModified: time.Now(),
	}, nil
}

// extractFrontmatter extracts YAML frontmatter from markdown content.
func extractFrontmatter(content string) (string, string, error) {
	// Match ---...--- at the start
	re := regexp.MustCompile(`(?s)^---\s*\n(.+?)\n---\s*\n(.*)$`)
	matches := re.FindStringSubmatch(content)

	if len(matches) != 3 {
		return "", content, nil // No frontmatter
	}

	return matches[1], matches[2], nil
}

// parseFrontmatter parses YAML frontmatter into metadata.
func parseFrontmatter(frontmatter string) (*SkillMetadata, error) {
	meta := &SkillMetadata{}

	lines := strings.Split(frontmatter, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "name":
			meta.Name = value
		case "description":
			meta.Description = value
		case "allowed-tools":
			// Parse comma-separated list
			if value != "" {
				tools := strings.Split(value, ",")
				for i, t := range tools {
					tools[i] = strings.TrimSpace(t)
				}
				meta.AllowedTools = tools
			}
		}
	}

	return meta, nil
}

// GetPrompt returns the skill content as a prompt.
func (s *Skill) GetPrompt(args string) string {
	prompt := s.Content

	// Replace {{args}} placeholder if present
	if strings.Contains(prompt, "{{args}}") {
		prompt = strings.ReplaceAll(prompt, "{{args}}", args)
	} else if args != "" {
		// Append args if no placeholder
		prompt = prompt + "\n\n## Arguments\n\n" + args
	}

	return prompt
}

// IsToolAllowed checks if a tool is allowed for this skill.
func (s *Skill) IsToolAllowed(toolName string) bool {
	// If no restrictions, all tools allowed
	if len(s.AllowedTools) == 0 {
		return true
	}

	// Check for wildcard
	for _, t := range s.AllowedTools {
		if t == "*" {
			return true
		}
	}

	// Check specific tool
	for _, t := range s.AllowedTools {
		if strings.EqualFold(t, toolName) {
			return true
		}
	}

	return false
}

// SkillInvocation represents an invocation of a skill.
type SkillInvocation struct {
	SkillName string    `json:"skill_name"`
	Args      string    `json:"args,omitempty"`
	StartedAt time.Time `json:"started_at"`
	Result    string    `json:"result,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// SkillContext provides context for skill execution.
type SkillContext struct {
	Ctx           context.Context
	WorkingDir    string
	ProjectRoot   string
	CurrentFile   string
	SelectionText string
	Variables     map[string]string
}

// NewSkillContext creates a new skill context.
func NewSkillContext(ctx context.Context, workingDir string) *SkillContext {
	return &SkillContext{
		Ctx:        ctx,
		WorkingDir: workingDir,
		Variables:  make(map[string]string),
	}
}

// ExpandVariables expands variables in the skill content.
func (sc *SkillContext) ExpandVariables(content string) string {
	// Built-in variables
	content = strings.ReplaceAll(content, "{{cwd}}", sc.WorkingDir)
	content = strings.ReplaceAll(content, "{{project_root}}", sc.ProjectRoot)
	content = strings.ReplaceAll(content, "{{current_file}}", sc.CurrentFile)
	content = strings.ReplaceAll(content, "{{selection}}", sc.SelectionText)

	// Custom variables
	for k, v := range sc.Variables {
		content = strings.ReplaceAll(content, "{{" + k + "}}", v)
	}

	return content
}
