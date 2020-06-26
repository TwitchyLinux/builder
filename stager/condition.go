package stager

import (
	"fmt"
	"reflect"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/pelletier/go-toml"
)

// StepCondition constrains a step based on conditionals.
type StepCondition struct {
	All []string `toml:"all"`
	Not []string `toml:"not"`
	Any []string `toml:"any"`
}

// ShouldSkip returns true if evaluation of the conditionals indicates
// that this block should be skipped.
func (c *StepCondition) ShouldSkip(tree *toml.Tree, opts Options) (bool, error) {
	if c == nil {
		return false, nil
	}
	env, err := cel.NewEnv(cel.Declarations(decls.NewIdent("conf", decls.Dyn, nil),
		decls.NewIdent("opts", decls.Dyn, nil),
		decls.NewIdent("features", decls.Dyn, nil)))
	if err != nil {
		return false, err
	}

	var f map[string]interface{}
	if features := tree.Get("features"); features != nil {
		switch ft := features.(type) {
		case *toml.Tree:
			f = ft.ToMap()
		case map[string]interface{}:
			f = ft
		case []string:
			f = make(map[string]interface{}, len(ft))
			for _, v := range ft {
				f[v] = true
			}
		}
	}

	m := map[string]interface{}{
		"conf":     tree.ToMap(),
		"opts":     opts,
		"features": f,
	}

	for i, e := range c.All {
		outcome, err := c.eval(env, e, m)
		if err != nil {
			return false, fmt.Errorf("evaluating if.all[%d]: %v", i, err)
		}
		if !outcome {
			return true, nil
		}
	}

	for i, e := range c.Not {
		outcome, err := c.eval(env, e, m)
		if err != nil {
			return false, fmt.Errorf("evaluating if.not[%d]: %v", i, err)
		}
		if outcome {
			return true, nil
		}
	}

	if len(c.Any) > 0 {
		for i, e := range c.Any {
			outcome, err := c.eval(env, e, m)
			if err != nil {
				return false, fmt.Errorf("evaluating if.any[%d]: %v", i, err)
			}
			if outcome {
				return false, nil
			}
		}
		return true, nil
	}
	return false, nil
}

func (c *StepCondition) eval(env *cel.Env, expr string, m map[string]interface{}) (bool, error) {
	parsed, issues := env.Parse(expr)
	if issues != nil && issues.Err() != nil {
		return false, issues.Err()
	}
	checked, issues := env.Check(parsed)
	if issues != nil && issues.Err() != nil {
		return false, issues.Err()
	}
	prg, err := env.Program(checked)
	if err != nil {
		return false, err
	}
	out, _, err := prg.Eval(m)
	if err != nil {
		return false, err
	}
	v, err := out.ConvertToNative(reflect.TypeOf(true))
	if err != nil {
		return false, err
	}
	return v.(bool), nil
}
