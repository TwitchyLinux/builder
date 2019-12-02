package user

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

var testGroup = `root:x:0:
daemon:x:1:
bin:x:2:
sys:x:3:
adm:x:4:
tty:x:5:
disk:x:6:
lp:x:7:
mail:x:8:
news:x:9:
uucp:x:10:
man:x:12:
proxy:x:13:
kmem:x:15:
dialout:x:20:
fax:x:21:
voice:x:22:
cdrom:x:24:
floppy:x:25:
tape:x:26:
sudo:x:27:
audio:x:29:
dip:x:30:
www-data:x:33:
backup:x:34:
operator:x:37:
list:x:38:
irc:x:39:
src:x:40:
gnats:x:41:
shadow:x:42:
utmp:x:43:
video:x:44:twl,yeet
sasl:x:45:
plugdev:x:46:
staff:x:50:
games:x:60:
users:x:100:
nogroup:x:65534:
systemd-journal:x:101:
systemd-timesync:x:102:
systemd-network:x:103:
systemd-resolve:x:104:
input:x:105:
kvm:x:106:
render:x:107:
crontab:x:108:
netdev:x:109:
Debian-exim:x:110:
ssh:x:111:
`

func TestParseGroup(t *testing.T) {
	find := func(entries []GroupEntry, name string) GroupEntry {
		for _, e := range entries {
			if e.Name == name {
				return e
			}
		}
		t.Errorf("could not find entry: %q", name)
		return GroupEntry{}
	}

	entries, err := ParseGroup(strings.NewReader(testGroup))
	if err != nil {
		t.Fatalf("ParseGroup() failed: %v", err)
	}

	if got, want := find(entries, "floppy"), (GroupEntry{
		Name: "floppy",
		Pass: "x",
		ID:   25,
	}); !reflect.DeepEqual(got, want) {
		t.Errorf("entries['floppy'] = %v, want %v", got, want)
	}
	if got, want := find(entries, "video"), (GroupEntry{
		Name:  "video",
		Pass:  "x",
		ID:    44,
		Users: []string{"twl", "yeet"},
	}); !reflect.DeepEqual(got, want) {
		t.Errorf("entries['video'] = %v, want %v", got, want)
	}
}

func TestGroupSerialize(t *testing.T) {
	entries, err := ParseGroup(strings.NewReader(testGroup))
	if err != nil {
		t.Fatal(err)
	}

	out, err := GroupSerialize(entries)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal([]byte(testGroup), out) {
		t.Errorf("group = %q, want %q", string(out), testGroup)
	}
}
