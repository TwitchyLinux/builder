package user

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type PasswdPassMode uint8

const (
	PassShadow PasswdPassMode = iota
	PassPlaintext
)

type PasswdPass struct {
	Mode PasswdPassMode
	Pass string
}

func (p *PasswdPass) String() string {
	switch p.Mode {
	case PassPlaintext:
		return p.Pass
	default:
		fallthrough
	case PassShadow:
		return "x"
	}
}

func passwdPass(pass string) PasswdPass {
	switch pass {
	case "x", "*":
		return PasswdPass{Mode: PassShadow}
	default:
		return PasswdPass{Mode: PassPlaintext, Pass: pass}
	}
}

// PasswdEntry represents a line in /etc/passwd.
type PasswdEntry struct {
	Username  string
	Password  PasswdPass
	UID, GID  int
	UserInfo  string
	HomeDir   string
	ShellPath string
}

func (e *PasswdEntry) String() string {
	var out strings.Builder
	out.WriteString(e.Username)
	out.WriteRune(':')
	out.WriteString(e.Password.String())
	out.WriteRune(':')

	out.WriteString(fmt.Sprint(e.UID))
	out.WriteRune(':')
	out.WriteString(fmt.Sprint(e.GID))
	out.WriteRune(':')

	out.WriteString(e.UserInfo)
	out.WriteRune(':')
	out.WriteString(e.HomeDir)
	out.WriteRune(':')
	out.WriteString(e.ShellPath)

	out.WriteRune('\n')
	return out.String()
}

func parsePasswdEntry(spl []string) (PasswdEntry, error) {
	entry := PasswdEntry{
		Username:  spl[0],
		Password:  passwdPass(spl[1]),
		UserInfo:  spl[4],
		HomeDir:   spl[5],
		ShellPath: spl[6],
	}
	var err error

	if entry.UID, err = strconv.Atoi(spl[2]); err != nil {
		return PasswdEntry{}, fmt.Errorf("uid: %v", err)
	}
	if entry.GID, err = strconv.Atoi(spl[3]); err != nil {
		return PasswdEntry{}, fmt.Errorf("gid: %v", err)
	}

	return entry, nil
}

// ParsePasswd parses content formatted as a /etc/passwd file.
func ParsePasswd(r io.Reader) ([]PasswdEntry, error) {
	var (
		s   = bufio.NewScanner(r)
		out []PasswdEntry
		i   int
	)

	for s.Scan() {
		spl := strings.Split(s.Text(), ":")
		if len(spl) < 7 {
			return nil, fmt.Errorf("line %d: expected >= 7 colon-separated elements, got %d", i+1, len(spl))
		}
		entry, err := parsePasswdEntry(spl)
		if err != nil {
			return nil, fmt.Errorf("line %d: %v", i+1, err)
		}

		i++
		out = append(out, entry)
	}
	if err := s.Err(); err != nil {
		return nil, fmt.Errorf("parsing passwd: %v", err)
	}

	return out, nil
}

// PasswdSerialize serializes the given user entries into the format expected
// at /etc/passwd.
func PasswdSerialize(entries []PasswdEntry) ([]byte, error) {
	var (
		out       bytes.Buffer
		dupeCheck = map[string]int{}
	)

	for i, e := range entries {
		if line, exists := dupeCheck[e.Username]; exists {
			return nil, fmt.Errorf("duplicate entry for user %q: entries %d & %d", e.Username, line, i)
		}
		dupeCheck[e.Username] = i

		out.WriteString(e.String())
	}

	return out.Bytes(), nil
}
