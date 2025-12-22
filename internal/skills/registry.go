package skills

import (
	"sort"
	"strings"
	"sync"
)

// SkillRegistry maintains a registry of available skills.
type SkillRegistry struct {
	mu     sync.RWMutex
	skills map[string]*Skill
}

// NewSkillRegistry creates a new skill registry.
func NewSkillRegistry() *SkillRegistry {
	return &SkillRegistry{
		skills: make(map[string]*Skill),
	}
}

// Register adds a skill to the registry.
// If a skill with the same name exists, it will be replaced.
func (r *SkillRegistry) Register(skill *Skill) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Normalize name to lowercase
	name := strings.ToLower(skill.Name)
	r.skills[name] = skill
}

// Get retrieves a skill by name.
func (r *SkillRegistry) Get(name string) *Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.skills[strings.ToLower(name)]
}

// GetByPrefix retrieves skills matching a prefix.
func (r *SkillRegistry) GetByPrefix(prefix string) []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	prefix = strings.ToLower(prefix)
	var matches []*Skill

	for name, skill := range r.skills {
		if strings.HasPrefix(name, prefix) {
			matches = append(matches, skill)
		}
	}

	// Sort by name
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Name < matches[j].Name
	})

	return matches
}

// List returns all registered skills.
func (r *SkillRegistry) List() []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skills := make([]*Skill, 0, len(r.skills))
	for _, skill := range r.skills {
		skills = append(skills, skill)
	}

	// Sort by name
	sort.Slice(skills, func(i, j int) bool {
		return skills[i].Name < skills[j].Name
	})

	return skills
}

// ListBySource returns skills from a specific source.
func (r *SkillRegistry) ListBySource(source string) []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var skills []*Skill
	for _, skill := range r.skills {
		if skill.Source == source {
			skills = append(skills, skill)
		}
	}

	sort.Slice(skills, func(i, j int) bool {
		return skills[i].Name < skills[j].Name
	})

	return skills
}

// Search searches for skills matching a query.
func (r *SkillRegistry) Search(query string) []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query = strings.ToLower(query)
	var matches []*Skill

	for _, skill := range r.skills {
		// Search in name and description
		if strings.Contains(strings.ToLower(skill.Name), query) ||
			strings.Contains(strings.ToLower(skill.Description), query) {
			matches = append(matches, skill)
		}
	}

	// Sort by relevance (exact name match first, then alphabetically)
	sort.Slice(matches, func(i, j int) bool {
		iExact := strings.ToLower(matches[i].Name) == query
		jExact := strings.ToLower(matches[j].Name) == query
		if iExact && !jExact {
			return true
		}
		if !iExact && jExact {
			return false
		}
		return matches[i].Name < matches[j].Name
	})

	return matches
}

// Remove removes a skill from the registry.
func (r *SkillRegistry) Remove(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	name = strings.ToLower(name)
	if _, exists := r.skills[name]; exists {
		delete(r.skills, name)
		return true
	}
	return false
}

// Count returns the number of registered skills.
func (r *SkillRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.skills)
}

// Clear removes all skills from the registry.
func (r *SkillRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.skills = make(map[string]*Skill)
}

// Exists checks if a skill exists.
func (r *SkillRegistry) Exists(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.skills[strings.ToLower(name)]
	return exists
}
