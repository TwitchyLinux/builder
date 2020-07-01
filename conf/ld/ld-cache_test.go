package ld

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var testUsingSystem = flag.Bool("test-system", false, "")

func TestParseCache(t *testing.T) {
	f, err := os.Open("ld.so.cache")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	out, err := ParseCache(f)
	if err != nil {
		t.Fatalf("ParseCache() failed: %v", err)
	}

	if got, want := out.Format, Glibc11Format; got != want {
		t.Errorf("format = %v, want %v", got, want)
	}
	if len(out.Entries) != 1270 {
		t.Fatalf("len(entries) = %v, want %v", len(out.Entries), 1270)
	}
}

func TestLookup(t *testing.T) {
	f, err := os.Open("ld.so.cache")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	out, err := ParseCache(f)
	if err != nil {
		t.Fatalf("ParseCache() failed: %v", err)
	}

	tcs := []struct {
		lookup, path string
		flags        EntryFlags
	}{
		{"libselinux.so.1", "/lib/x86_64-linux-gnu/libselinux.so.1", EntryFlags(PlatformX64)},
		{"libm.so.6", "/lib/x86_64-linux-gnu/libm.so.6", EntryFlags(PlatformX64)},
		{"libc.so.6", "/lib/x86_64-linux-gnu/libc.so.6", EntryFlags(PlatformX64)},
		{"libpcre.so.3", "/lib/x86_64-linux-gnu/libpcre.so.3", EntryFlags(PlatformX64)},
		{"libpthread.so.0", "/lib/x86_64-linux-gnu/libpthread.so.0", EntryFlags(PlatformX64)},
		{"", "", 0},
	}

	for _, tc := range tcs {
		t.Run(tc.lookup, func(t *testing.T) {
			got, want := out.Lookup(tc.lookup, PlatformX64), &CacheEntry{
				Flags: tc.flags,
				Key:   filepath.Base(tc.path),
				Val:   tc.path,
			}
			if tc.path == "" {
				if got != nil {
					t.Errorf("Lookup(%q) = %+v, want nil", tc.lookup, got)
				}
			} else if *got != *want {
				t.Errorf("Lookup(%q) = %+v, want %+v", tc.lookup, got, want)
			}
		})
	}

}

func TestParseSystemCache(t *testing.T) {
	if !*testUsingSystem {
		t.SkipNow()
	}

	f, err := os.Open("/etc/ld.so.cache")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if _, err := ParseCache(f); err != nil {
		t.Errorf("ParseCache('/etc/ld.so.cache') failed: %v", err)
	}
}
