package skills

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// SkillInvoker handles skill invocation.
type SkillInvoker struct {
	registry *SkillRegistry
	loader   *SkillLoader
	history  []SkillInvocation
}

// NewSkillInvoker creates a new skill invoker.
func NewSkillInvoker(registry *SkillRegistry, loader *SkillLoader) *SkillInvoker {
	return &SkillInvoker{
		registry: registry,
		loader:   loader,
		history:  []SkillInvocation{},
	}
}

// InvokeResult represents the result of invoking a skill.
type InvokeResult struct {
	Skill       *Skill    `json:"skill"`
	Prompt      string    `json:"prompt"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
}

// Invoke invokes a skill by name with arguments.
func (i *SkillInvoker) Invoke(ctx context.Context, skillName, args string) (*InvokeResult, error) {
	startedAt := time.Now()

	// Parse skill name (may include namespace like "namespace:skill")
	skillName = normalizeSkillName(skillName)

	// Get skill from registry
	skill := i.registry.Get(skillName)
	if skill == nil {
		// Try to load it
		var err error
		skill, err = i.loader.LoadSkillByName(skillName)
		if err != nil {
			return nil, fmt.Errorf("skill not found: %s", skillName)
		}
	}

	// Create skill context
	skillCtx := NewSkillContext(ctx, ".")

	// Expand variables in content
	expandedContent := skillCtx.ExpandVariables(skill.Content)

	// Build prompt
	prompt := buildSkillPrompt(skill, expandedContent, args)

	// Record invocation
	invocation := SkillInvocation{
		SkillName: skillName,
		Args:      args,
		StartedAt: startedAt,
	}
	i.history = append(i.history, invocation)

	return &InvokeResult{
		Skill:       skill,
		Prompt:      prompt,
		StartedAt:   startedAt,
		CompletedAt: time.Now(),
	}, nil
}

// normalizeSkillName normalizes a skill name.
func normalizeSkillName(name string) string {
	// Remove leading slash if present
	name = strings.TrimPrefix(name, "/")

	// Handle namespace:skill format
	if strings.Contains(name, ":") {
		parts := strings.SplitN(name, ":", 2)
		if len(parts) == 2 {
			// Return namespace-skill format
			return parts[0] + "-" + parts[1]
		}
	}

	return name
}

// buildSkillPrompt builds the prompt for skill execution.
func buildSkillPrompt(skill *Skill, content, args string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Skill: %s\n\n", skill.Name))

	if skill.Description != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", skill.Description))
	}

	if len(skill.AllowedTools) > 0 {
		sb.WriteString(fmt.Sprintf("**Allowed Tools**: %s\n\n", strings.Join(skill.AllowedTools, ", ")))
	}

	sb.WriteString("---\n\n")
	sb.WriteString(content)

	if args != "" {
		sb.WriteString("\n\n---\n\n")
		sb.WriteString(fmt.Sprintf("## Arguments\n\n%s", args))
	}

	return sb.String()
}

// GetHistory returns the invocation history.
func (i *SkillInvoker) GetHistory() []SkillInvocation {
	return i.history
}

// ClearHistory clears the invocation history.
func (i *SkillInvoker) ClearHistory() {
	i.history = []SkillInvocation{}
}

// GetLastInvocation returns the last invocation.
func (i *SkillInvoker) GetLastInvocation() *SkillInvocation {
	if len(i.history) == 0 {
		return nil
	}
	return &i.history[len(i.history)-1]
}
