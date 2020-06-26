package stager

import (
	"bytes"
	"fmt"

	"github.com/pelletier/go-toml"
	"github.com/twitchylinux/builder/conf/udev"
	"github.com/twitchylinux/builder/units"
)

const (
	udevNumberingStart = 52
)

// UdevRules describes a collection of udev rules.
type UdevRules struct {
	Name  string         `toml:"name"`
	If    *StepCondition `toml:"if"`
	Rules []UdevRule     `toml:"rules"`
}

// UdevRule describes a udev rule.
type UdevRule struct {
	Comment string       `toml:"comment"`
	If      []UdevMatch  `toml:"if"`
	Then    []UdevAction `toml:"then"`
}

// UdevMatch describes a condition applied to a udev rule.
type UdevMatch struct {
	Key string `toml:"key"`
	Val string `toml:"value"`
	Op  string `toml:"op" default:"=="`
}

// UdevAction describes an action to undertake if a udev rule matches.
type UdevAction struct {
	Action string `toml:"action"`
	Val    string `toml:"value"`
	Op     string `toml:"op" default:"="`
}

func makeUdevRule(r UdevRule) *udev.Rule {
	rule := udev.Rule{
		LeadingComment: r.Comment,
		Matches:        make([]udev.Match, len(r.If)),
		Actions:        make([]udev.Action, len(r.Then)),
	}
	for i := range r.If {
		rule.Matches[i] = udev.Match{
			Key: r.If[i].Key,
			Val: r.If[i].Val,
			Op:  udev.MatchOp(r.If[i].Op),
		}
	}
	for i := range r.Then {
		rule.Actions[i] = udev.Action{
			Key: r.Then[i].Action,
			Val: r.Then[i].Val,
			Op:  udev.ActionOp(r.Then[i].Op),
		}
	}
	return &rule
}

func udevConf(opts Options, tree *toml.Tree) (*units.InstallFiles, error) {
	conf := map[string]UdevRules{}
	t := tree.Get(keyUdevRules)
	if t == nil {
		return nil, nil
	}
	ge, ok := t.(*toml.Tree)
	if !ok {
		return nil, fmt.Errorf("invalid config: %s is not a structure (got %T)", keyUdevRules, t)
	}
	if err := ge.Unmarshal(&conf); err != nil {
		return nil, err
	}
	if len(conf) == 0 {
		return nil, nil
	}

	var (
		outFiles []units.FileInfo
		i        int
	)
	for name, ruleSet := range conf {
		skip, err := ruleSet.If.ShouldSkip(tree, opts)
		if err != nil {
			return nil, err
		}
		if skip {
			continue
		}

		var ruleContents bytes.Buffer
		for i, r := range ruleSet.Rules {
			if err := makeUdevRule(r).Serialize(&ruleContents); err != nil {
				return nil, fmt.Errorf("%v.%v[%d] failed serialization: %v", keyUdevRules, name, i, err)
			}
		}
		outFiles = append(outFiles, units.FileInfo{
			Path: fmt.Sprintf("/etc/udev/rules.d/%.2d-%s.rules", udevNumberingStart+i, name),
			Data: ruleContents.Bytes(),
		})
		i++
	}

	if len(outFiles) == 0 {
		return nil, nil
	}
	return &units.InstallFiles{
		UnitName: "udev-rules",
		Mkdir:    "/etc/udev/rules.d",
		Files:    outFiles,
	}, nil
}
