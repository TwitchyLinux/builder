package udev

import (
	"bytes"
	"testing"
)

func TestSerialize(t *testing.T) {
	var out bytes.Buffer
	rules := []Rule{
		{
			LeadingComment: "ello",
			Matches: []Match{
				{
					Op:  Equal,
					Key: MatchSubsystem,
					Val: "tty",
				},
				{
					Op:  Equal,
					Key: MatchAnyProductID,
					Val: "6001",
				},
			},
			Actions: []Action{
				{
					Op:  Assign,
					Key: ActionGroup,
					Val: "users",
				},
				{
					Op:  Assign,
					Key: ActionMode,
					Val: "0666",
				},
				{
					Op:  Append,
					Key: ActionSymlink,
					Val: "buspirate",
				},
				{
					Op:     Assign,
					Key:    ActionEnv,
					Subkey: "ID_MM_DEVICE_IGNORE",
					Val:    "1",
				},
			},
		},
	}
	for i, r := range rules {
		if err := r.Serialize(&out); err != nil {
			t.Errorf("rule at index %d failed to serialize: %v", i, err)
		}
	}

	want := "# ello\nSUBSYSTEM==\"tty\", ATTRS{idProduct}==\"6001\", GROUP=\"users\", MODE=\"0666\", SYMLINK+=\"buspirate\", ENV{ID_MM_DEVICE_IGNORE}=\"1\"\n"
	if want != out.String() {
		t.Errorf("got = %q, want %q", out.String(), want)
	}
}
