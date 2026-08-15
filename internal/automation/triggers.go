package automation

import (
	"regexp"
	"strings"
	"sync"

	"vibeterm/internal/models"
)

// TriggerEngine monitors terminal stream patterns and triggers macros or automated responses
type TriggerEngine struct {
	mu    sync.RWMutex
	rules []models.TriggerRule
}

// NewTriggerEngine creates a new TriggerEngine
func NewTriggerEngine() *TriggerEngine {
	return &TriggerEngine{
		rules: []models.TriggerRule{
			{
				ID:      "sudo-prompt",
				Pattern: `\[sudo\] password for `,
				IsRegex: true,
				Action:  "highlight",
				Payload: "#FF5555",
				Enabled: true,
			},
			{
				ID:      "cisco-enable-prompt",
				Pattern: `Password:`,
				IsRegex: false,
				Action:  "highlight",
				Payload: "#50FA7B",
				Enabled: true,
			},
		},
	}
}

// CheckMatch checks a line of text against active trigger rules
func (te *TriggerEngine) CheckMatch(line string) []models.TriggerRule {
	te.mu.RLock()
	defer te.mu.RUnlock()

	var matches []models.TriggerRule
	for _, rule := range te.rules {
		if !rule.Enabled {
			continue
		}
		if rule.IsRegex {
			if matched, _ := regexp.MatchString(rule.Pattern, line); matched {
				matches = append(matches, rule)
			}
		} else {
			if strings.Contains(line, rule.Pattern) {
				matches = append(matches, rule)
			}
		}
	}
	return matches
}
