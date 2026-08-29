package automation

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type MacroStepType string

const (
	StepSend       MacroStepType = "send"
	StepWaitFor    MacroStepType = "wait_for"
	StepSleep      MacroStepType = "sleep"
	StepAssertExit MacroStepType = "assert_exit"
)

type MacroStep struct {
	Type    MacroStepType `json:"type"`
	Payload string        `json:"payload"`
	Timeout int           `json:"timeoutMs"` // milliseconds
}

type MacroPlaybook struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Variables   map[string]string `json:"variables"`
	Steps       []MacroStep       `json:"steps"`
}

// SubstituteVariables replaces {{var}} with matching values
func SubstituteVariables(input string, vars map[string]string) string {
	result := input
	for k, v := range vars {
		placeholder := fmt.Sprintf("{{%s}}", k)
		result = strings.ReplaceAll(result, placeholder, v)
	}
	return result
}

// MacroRunner executes an automation playbook against a terminal stream
type MacroRunner struct{}

func NewMacroRunner() *MacroRunner {
	return &MacroRunner{}
}

func (mr *MacroRunner) RunStep(ctx context.Context, step MacroStep, sendFunc func(string) error, vars map[string]string) error {
	payload := SubstituteVariables(step.Payload, vars)

	switch step.Type {
	case StepSend:
		return sendFunc(payload + "\n")
	case StepSleep:
		ms := 500
		if step.Timeout > 0 {
			ms = step.Timeout
		}
		time.Sleep(time.Duration(ms) * time.Millisecond)
		return nil
	default:
		return nil
	}
}
