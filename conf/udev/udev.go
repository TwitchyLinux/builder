// Package udev serializes udev rules.
package udev

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"strings"
)

// Rule describes a udev rule.
type Rule struct {
	LeadingComment string
	Matches        []Match
	Actions        []Action
}

// Serialize generates formatted udev rules.
func (r *Rule) Serialize(w io.Writer) error {
	if len(r.Matches) == 0 {
		return errors.New("udev rules must contain match elements")
	}
	if len(r.Actions) == 0 {
		return errors.New("udev rules must contain action elements")
	}
	out := bufio.NewWriter(w)

	if r.LeadingComment != "" {
		out.WriteString("# " + strings.Replace(r.LeadingComment, "\n", " ", -1) + "\n")
	}

	var elements []string
	for _, m := range r.Matches {
		elements = append(elements, m.Key+string(m.Op)+strconv.Quote(m.Val))
	}
	for _, a := range r.Actions {
		elements = append(elements, a.Key+string(a.Op)+strconv.Quote(a.Val))
	}

	out.WriteString(strings.Join(elements, ", ") + "\n")
	return out.Flush()
}

// ActionOp describes how the action should be performed.
type ActionOp string

// Valid ActionOp values.
const (
	Assign ActionOp = "="
	Append ActionOp = "+="
)

// Action describes an action to be performed if the rule matches.
type Action struct {
	Op       ActionOp
	Key, Val string
}

// MatchOp describes how the match key will be checked against the value.
type MatchOp string

// Valid MatchOp values.
const (
	Equal    MatchOp = "=="
	NotEqual MatchOp = "!="
)

type Match struct {
	Op       MatchOp
	Key, Val string
}
