package units

import (
	"runtime"
	"testing"
)

func TestFindBinary(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.SkipNow()
	}

	p, err := FindBinary("bash")
	if err != nil {
		t.Fatal(err)
	}
	switch p {
	case "/bin/bash", "/usr/bin/bash", "/sbin/bash":
	default:
		t.Errorf("Unexpected path: %v", p)
	}
}
