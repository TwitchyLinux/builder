package dconf

import "bytes"

// LocksDir specifies the relative path to lock configuration for
// a system database.
const LocksDir = "locks"

// Lock specifies a configuration key in a system database that
// overrides other, higher-level keys (such as in a user db).
type Lock string

// Generate will create the contents of the lock file, to be
// placed at <db-dir>/locks/<name>.
func (l Lock) Generate() []byte {
	var b bytes.Buffer
	b.WriteString(string(l))
	b.WriteRune('\n')
	return b.Bytes()
}
