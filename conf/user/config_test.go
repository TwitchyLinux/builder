package user

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestReadAdduser(t *testing.T) {
	c := Config{}
	expect := map[string]string{
		"DHOME":            "/home",
		"LAST_UID":         "59999",
		"FIRST_UID":        "1000",
		"LAST_GID":         "59999",
		"FIRST_GID":        "1000",
		"LAST_SYSTEM_UID":  "999",
		"FIRST_SYSTEM_UID": "100",
		"LAST_SYSTEM_GID":  "999",
		"FIRST_SYSTEM_GID": "100",
	}

	out, err := c.readAdduserConf("testdata/adduser.conf")
	if err != nil {
		t.Fatalf("readAdduserConf() failed: %v", err)
	}

	for k, want := range expect {
		if got := out[k]; got != want {
			t.Errorf("conf[%q] = %q, want %q", k, got, want)
		}
	}
}

func TestUpsertUser(t *testing.T) {
	now := time.Now().UTC()
	tcs := []struct {
		name     string
		passwd   string
		group    string
		adduser  string
		newUsers []string
		users    []PasswdEntry
		groups   []GroupEntry
		shadow   []ShadowEntry
	}{
		{
			name: "no-config",
		},
		{
			name:     "defaults",
			newUsers: []string{"test"},
			users:    []PasswdEntry{{Username: "test", UID: 1000, GID: 1000, HomeDir: "/home/test", ShellPath: "/bin/bash"}},
			groups:   []GroupEntry{{Name: "test", Pass: "x", ID: 1000}},
			shadow:   []ShadowEntry{{Username: "test", LastChanged: now, MaxChangeDays: 99999, WarnBeforeMaxDays: 7}},
		},
		{
			name:     "defaults-uid-already-exists",
			passwd:   (&PasswdEntry{Username: "exists", UID: 1000, GID: 100}).String(),
			newUsers: []string{"test"},
			users: []PasswdEntry{
				{Username: "exists", UID: 1000, GID: 100},
				{Username: "test", UID: 1001, GID: 1001, HomeDir: "/home/test", ShellPath: "/bin/bash"},
			},
			groups: []GroupEntry{{Name: "test", Pass: "x", ID: 1001}},
			shadow: []ShadowEntry{{Username: "test", LastChanged: now, MaxChangeDays: 99999, WarnBeforeMaxDays: 7}},
		},
		{
			name:     "defaults-gid-already-exists",
			group:    (&GroupEntry{Name: "exists", ID: 1000}).String(),
			newUsers: []string{"test"},
			users: []PasswdEntry{
				{Username: "test", UID: 1000, GID: 1001, HomeDir: "/home/test", ShellPath: "/bin/bash"},
			},
			groups: []GroupEntry{
				{Name: "exists", ID: 1000},
				{Name: "test", Pass: "x", ID: 1001},
			},
			shadow: []ShadowEntry{{Username: "test", LastChanged: now, MaxChangeDays: 99999, WarnBeforeMaxDays: 7}},
		},
		{
			name:     "upsert",
			passwd:   (&PasswdEntry{Username: "test", UID: 123, GID: 123}).String(),
			newUsers: []string{"test"},
			users:    []PasswdEntry{{Username: "test", UID: 123, GID: 123}},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			root, cleanup := makeTestRoot(t, tc.passwd, tc.group, tc.adduser)
			defer cleanup()
			c, err := ReadConfig(root)
			if err != nil {
				t.Errorf("ReadConfig() failed: %v", err)
			}
			c.now = now

			for _, name := range tc.newUsers {
				if err := c.UpsertUser(name); err != nil {
					t.Errorf("UpsertUser(%q) failed: %v", name, err)
				}
			}

			if got, want := c.users, tc.users; !reflect.DeepEqual(got, want) {
				t.Errorf("c.users = %v, want %v", got, want)
			}
			if got, want := c.groups, tc.groups; !reflect.DeepEqual(got, want) {
				t.Errorf("c.groups = %v, want %v", got, want)
			}
			if got, want := c.shadow, tc.shadow; !reflect.DeepEqual(got, want) {
				t.Errorf("c.shadow = %v, want %v", got, want)
			}
		})
	}
}

func makeTestRoot(t *testing.T, passwd, group, adduser string) (string, func()) {
	t.Helper()
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "etc"), 0755); err != nil {
		t.Fatal(err)
	}

	if passwd != "" {
		if err := ioutil.WriteFile(filepath.Join(dir, "etc", "passwd"), []byte(passwd), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if group != "" {
		if err := ioutil.WriteFile(filepath.Join(dir, "etc", "group"), []byte(group), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if adduser != "" {
		if err := ioutil.WriteFile(filepath.Join(dir, "etc", "adduser.conf"), []byte(adduser), 0644); err != nil {
			t.Fatal(err)
		}
	}

	return dir, func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Errorf("failed to clean up test: %v", err)
		}
	}
}
