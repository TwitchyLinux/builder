package units

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Masterminds/semver"
)

var (
	vers         = regexp.MustCompile("([0-9]+\\.?)+")
	versTrailing = regexp.MustCompile("([0-9]+\\.?)+$")
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

// CmdOutput returns the full standard output of invoking the given binary
// given the provided arguments.
func CmdOutput(ctx context.Context, bin string, args ...string) (string, error) {
	p, err := FindBinary(bin)
	if err != nil {
		return "", fmt.Errorf("%s: %v", bin, err)
	}
	cmd := exec.CommandContext(ctx, p)
	cmd.Args = append([]string{p}, args...)
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	return string(out), err
}

// CmdCombined returns the full standard output and standard error for the
// invocation of the provided binary given the provided arguments.
func CmdCombined(ctx context.Context, bin string, args ...string) (string, error) {
	p, err := FindBinary(bin)
	if err != nil {
		return "", fmt.Errorf("%s: %v", bin, err)
	}
	cmd := exec.CommandContext(ctx, p)
	cmd.Args = append([]string{p}, args...)

	out, err := cmd.CombinedOutput()
	return string(out), err
}

// CompareExtractSemver extracts a semver from the version string,
// returning the result of a version comparison with wantVersion.
func CompareExtractSemver(version, wantVersion string) (int, error) {
	version = strings.TrimSpace(version)
	s := extractSemver(version, versTrailing)
	if s == "" {
		s = extractSemver(version, vers)
	}
	if s == "" {
		return 0, errors.New("failed to determine version")
	}

	v, err := semver.NewVersion(s)
	if err != nil {
		return 0, err
	}
	wantV, err := semver.NewVersion(wantVersion)
	if err != nil {
		return 0, err
	}
	c := wantV.Compare(v)
	return c, nil
}

func extractSemver(version string, regex *regexp.Regexp) string {
	for _, candidate := range strings.Split(version, " ") {
		if regex.MatchString(candidate) {
			var out strings.Builder
			for _, c := range candidate {
				if strings.ContainsAny(string(c), "0123456789.") {
					out.WriteRune(c)
				} else {
					break
				}
			}
			return out.String()
		}
	}
	return ""
}
