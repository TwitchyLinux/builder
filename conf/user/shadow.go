package user

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/tredoe/osutil/user/crypt"
)

type ShadowPassMode uint8

// Valid password modes.
const (
	PassEncrypted ShadowPassMode = iota
	PassNotRequired
	PassAccountDisabled
)

// ShadowPass represents the shadow configuration for a password field.
type ShadowPass struct {
	Mode      ShadowPassMode
	Encrypted string
}

func (p *ShadowPass) String() string {
	switch p.Mode {
	case PassNotRequired:
		return ""
	case PassEncrypted:
		return p.Encrypted
	default:
		fallthrough
	case PassAccountDisabled:
		return "*"
	}
}

// ShadowHash computes the hash of the given password, returning a form which
// can be written into a shadow file.
func ShadowHash(pw string) (ShadowPass, error) {
	encrypted, err := crypt.New(crypt.SHA512).Generate([]byte(pw), nil)
	if err != nil {
		return ShadowPass{}, nil
	}
	return ShadowPass{
		Mode:      PassEncrypted,
		Encrypted: encrypted,
	}, nil
}

func shadowPass(pass string) ShadowPass {
	switch pass {
	case "":
		return ShadowPass{Mode: PassNotRequired}
	case "*", "!":
		return ShadowPass{Mode: PassAccountDisabled}
	default:
		return ShadowPass{Mode: PassEncrypted, Encrypted: pass}
	}
}

// ShadowEntry describes an entry in /etc/shadow.
type ShadowEntry struct {
	Username string
	Password ShadowPass

	LastChanged time.Time
	Expiry      time.Time

	MinChangeDays          int
	MaxChangeDays          int
	WarnBeforeMaxDays      int
	DisableAfterExpiryDays int
}

func (e *ShadowEntry) String() string {
	var out strings.Builder
	out.WriteString(e.Username)
	out.WriteRune(':')
	out.WriteString(e.Password.String())
	out.WriteRune(':')

	out.WriteString(fmt.Sprint(toEpochDays(e.LastChanged)))
	out.WriteRune(':')
	out.WriteString(fmt.Sprint(e.MinChangeDays))
	out.WriteRune(':')
	out.WriteString(fmt.Sprint(e.MaxChangeDays))
	out.WriteRune(':')
	out.WriteString(fmt.Sprint(e.WarnBeforeMaxDays))
	out.WriteRune(':')
	if e.DisableAfterExpiryDays > 0 {
		out.WriteString(fmt.Sprint(e.DisableAfterExpiryDays))
	}
	out.WriteRune(':')
	if !e.Expiry.IsZero() {
		out.WriteString(fmt.Sprint(toEpochDays(e.Expiry)))
	}
	out.WriteRune(':')

	out.WriteRune('\n')
	return out.String()
}

func fromEpochDays(days int) time.Time {
	return time.Unix(int64(days)*86400, 0).UTC()
}

func toEpochDays(t time.Time) int {
	return int(t.UTC().Unix() / 86400)
}

func convertEpochDays(days string) (time.Time, error) {
	d, err := strconv.Atoi(days)
	if err != nil {
		return time.Time{}, err
	}
	return fromEpochDays(d), nil
}

func parseShadowEntry(spl []string) (ShadowEntry, error) {
	entry := ShadowEntry{
		Username: spl[0],
		Password: shadowPass(spl[1]),
	}
	var err error

	if entry.LastChanged, err = convertEpochDays(spl[2]); err != nil {
		return ShadowEntry{}, fmt.Errorf("last changed: %v", err)
	}
	if entry.MinChangeDays, err = strconv.Atoi(spl[3]); err != nil {
		return ShadowEntry{}, fmt.Errorf("min days before change: %v", err)
	}
	if entry.MaxChangeDays, err = strconv.Atoi(spl[4]); err != nil {
		return ShadowEntry{}, fmt.Errorf("max days before mandatory change: %v", err)
	}
	if spl[5] != "" {
		if entry.WarnBeforeMaxDays, err = strconv.Atoi(spl[5]); err != nil {
			return ShadowEntry{}, fmt.Errorf("days before expiry warning: %v", err)
		}
	}
	if spl[6] != "" {
		if entry.DisableAfterExpiryDays, err = strconv.Atoi(spl[6]); err != nil {
			return ShadowEntry{}, fmt.Errorf("days after expiry disable: %v", err)
		}
	}
	if spl[7] != "" {
		if entry.Expiry, err = convertEpochDays(spl[7]); err != nil {
			return ShadowEntry{}, fmt.Errorf("expiry: %v", err)
		}
	}

	return entry, nil
}

// ParseShadow parses content formatted as a /etc/shadow file.
func ParseShadow(r io.Reader) ([]ShadowEntry, error) {
	var (
		s   = bufio.NewScanner(r)
		out []ShadowEntry
		i   int
	)

	for s.Scan() {
		spl := strings.Split(s.Text(), ":")
		if len(spl) < 9 {
			return nil, fmt.Errorf("line %d: expected >= 9 colon-separated elements, got %d", i+1, len(spl))
		}
		entry, err := parseShadowEntry(spl)
		if err != nil {
			return nil, fmt.Errorf("line %d: %v", i+1, err)
		}

		i++
		out = append(out, entry)
	}
	if err := s.Err(); err != nil {
		return nil, fmt.Errorf("parsing shadow: %v", err)
	}

	return out, nil
}

func ShadowEntryValidate(entry ShadowEntry) error {
	// Looks like this is no longer a thing.
	// if len(entry.Username) > 13 {
	// 	return errors.New("username exceeds 13 character max length")
	// }
	return nil
}

// ShadowSerialize serializes the given shadow entries into the format expected
// at /etc/shadow.
func ShadowSerialize(entries []ShadowEntry) ([]byte, error) {
	var (
		out       bytes.Buffer
		dupeCheck = map[string]int{}
	)

	for i, e := range entries {
		if line, exists := dupeCheck[e.Username]; exists {
			return nil, fmt.Errorf("duplicate entry for user %q: entries %d & %d", e.Username, line, i)
		}
		dupeCheck[e.Username] = i

		if err := ShadowEntryValidate(e); err != nil {
			return nil, fmt.Errorf("entry %d: %v", i, err)
		}
		out.WriteString(e.String())
	}

	return out.Bytes(), nil
}
