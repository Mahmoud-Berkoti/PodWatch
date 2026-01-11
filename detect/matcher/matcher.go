package matcher

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/podwatch/podwatch/pkg/models"
)

type RuleEngine struct {
	rules    []models.Rule
	programs map[string]cel.Program
	mu       sync.RWMutex
	env      *cel.Env
}

func NewRuleEngine(rules []models.Rule) (*RuleEngine, error) {
	// Define the environment
	// We expose "event" as a map for flexibility
	env, err := cel.NewEnv(
		cel.Declarations(
			decls.NewVar("event", decls.NewMapType(decls.String, decls.Dyn)),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL env: %w", err)
	}

	re := &RuleEngine{
		rules:    rules,
		programs: make(map[string]cel.Program),
		env:      env,
	}

	if err := re.CompileAll(); err != nil {
		return nil, err
	}

	return re, nil
}

func (re *RuleEngine) CompileAll() error {
	re.mu.Lock()
	defer re.mu.Unlock()

	for _, rule := range re.rules {
		if !rule.Enabled {
			continue
		}
		ast, issues := re.env.Compile(rule.Condition)
		if issues != nil && issues.Err() != nil {
			return fmt.Errorf("rule %s compile error: %w", rule.Name, issues.Err())
		}
		prg, err := re.env.Program(ast)
		if err != nil {
			return fmt.Errorf("rule %s program error: %w", rule.Name, err)
		}
		re.programs[rule.ID] = prg
	}
	return nil
}

func (re *RuleEngine) Evaluate(event models.RuntimeEvent) ([]models.Alert, error) {
	re.mu.RLock()
	defer re.mu.RUnlock()

	// Convert event to map
	// TODO: Optimization: Use reflection or struct directly if registered
	data, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}
	var inputMap map[string]interface{}
	if err := json.Unmarshal(data, &inputMap); err != nil {
		return nil, err
	}

	input := map[string]interface{}{
		"event": inputMap,
	}

	var alerts []models.Alert

	for _, rule := range re.rules {
		if !rule.Enabled {
			continue
		}
		prg, ok := re.programs[rule.ID]
		if !ok {
			continue
		}

		out, _, err := prg.Eval(input)
		if err != nil {
			log.Printf("Rule %s eval error: %v", rule.Name, err)
			continue
		}

		if out == nil {
			continue
		}

		match, ok := out.Value().(bool)
		if ok && match {
			alerts = append(alerts, models.Alert{
				// ID generated later
				RuleName:    rule.Name,
				Severity:    rule.Severity,
				Description: rule.Description,
				Event:       &event,
				Response:    rule.Response,
			})
		}
	}

	return alerts, nil
}
