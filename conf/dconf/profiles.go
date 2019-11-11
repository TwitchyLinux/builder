package dconf

import "bytes"

const (
	// ProfileDir is the default directory where profiles are defined.
	ProfileDir = "etc/dconf/profile"

	// User relates a user database at $XDG_CONFIG_HOME/dconf/<name>.
	User DirectiveType = "user-db"
	// System specifies that a system database should be read. The binary
	// database is read from /etc/dconf/db/<name>.
	System DirectiveType = "system-db"
	// Service relates a binary and text database pair. The binary database file
	// is installed at $XDG_RUNTIME_DIR, while the text database is stored at
	// $XDG_CONFIG_HOME/dconf/<name>.txt. The two files are kept up to date.
	Service DirectiveType = "service-db"
	// File relates a database to the file at the path specified by name.
	File DirectiveType = "file-db"
)

// DirectiveType how a database relates to a profile.
type DirectiveType string

// Directive relates a profile to a database.
type Directive struct {
	Type DirectiveType
	Name string
}

// Profile represents the configuration of a Dconf profile.
type Profile struct {
	RW  Directive
	ROs []Directive
}

// Generate serializes a profile into a format writeable to /etc/dconf/profile.
func (p *Profile) Generate() []byte {
	var b bytes.Buffer

	for _, d := range append([]Directive{p.RW}, p.ROs...) {
		b.WriteString(string(d.Type))
		b.WriteRune(':')
		b.WriteString(d.Name)
		b.WriteRune('\n')
	}
	return b.Bytes()
}
