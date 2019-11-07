package units

import (
	"os"
	"path/filepath"
	"strings"
)

func binarySearchPaths() []string {
	return strings.Split(os.Getenv("PATH"), ":")
}

// FindBinary returns the full path to the binary specified by resolving paths
// registered in the PATH environment variable.
func FindBinary(bin string) (string, error) {
	for _, p := range binarySearchPaths() {
		s, err := os.Stat(filepath.Join(p, bin))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return "", err
		}
		if s.Mode()&0111 != 0 {
			return filepath.Join(p, bin), nil
		}
	}
	return "", os.ErrNotExist
}
