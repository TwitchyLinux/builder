package ld

import (
	"os"
	"testing"
)

func TestParseCache(t *testing.T) {
	f, err := os.Open("ld.so.cache")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	out, err := ParseCache(f)
	if err != nil {
		t.Errorf("ParseCache() failed: %v", err)
	}

	if got, want := out.Format, Glibc11Format; got != want {
		t.Errorf("format = %v, want %v", got, want)
	}
	if len(out.Entries) != 1270 {
		t.Fatalf("len(entries) = %v, want %v", len(out.Entries), 1270)
	}
}
