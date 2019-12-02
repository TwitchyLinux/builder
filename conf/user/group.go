package user

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// GroupEntry describes a group in an /etc/group file.
type GroupEntry struct {
	Name  string
	Pass  string
	ID    int
	Users []string
}

func (g GroupEntry) String() string {
	return strings.Join([]string{g.Name, g.Pass, strconv.Itoa(g.ID), strings.Join(g.Users, ",")}, ":") + "\n"
}

func parseGroupEntry(spl []string) (GroupEntry, error) {
	group := GroupEntry{
		Name: spl[0],
		Pass: spl[1],
	}
	if len(strings.TrimSpace(spl[3])) > 0 {
		group.Users = strings.Split(spl[3], ",")
	}

	var err error
	group.ID, err = strconv.Atoi(spl[2])
	return group, err
}

// ParseGroup parses content formatted as a /etc/group file.
func ParseGroup(r io.Reader) ([]GroupEntry, error) {
	var (
		s   = bufio.NewScanner(r)
		out []GroupEntry
		i   int
	)

	for s.Scan() {
		spl := strings.Split(s.Text(), ":")
		if len(spl) != 4 {
			return nil, fmt.Errorf("line %d: expected 4 colon-separated elements, got %d", i+1, len(spl))
		}
		entry, err := parseGroupEntry(spl)
		if err != nil {
			return nil, fmt.Errorf("line %d: %v", i+1, err)
		}

		i++
		out = append(out, entry)
	}
	if err := s.Err(); err != nil {
		return nil, fmt.Errorf("parsing group: %v", err)
	}

	return out, nil
}

// GroupSerialize serializes the given group entries into the format expected
// at /etc/group.
func GroupSerialize(entries []GroupEntry) ([]byte, error) {
	var (
		out       bytes.Buffer
		dupeCheck = map[string]int{}
	)

	for i, e := range entries {
		if line, exists := dupeCheck[e.Name]; exists {
			return nil, fmt.Errorf("duplicate entry for group %q: entries %d & %d", e.Name, line, i)
		}
		dupeCheck[e.Name] = i

		out.WriteString(e.String())
	}

	return out.Bytes(), nil
}
