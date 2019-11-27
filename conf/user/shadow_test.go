package user

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
	"time"
)

var testShadow = `root:*:18226:0:99999:7:::
daemon:*:18226:0:99999:7:::
bin:*:18226:0:99999:7:::
sys:*:18226:0:99999:7:::
sync:*:18226:0:99999:7:::
games:*:18226:0:99999:7:::
man:*:18226:0:99999:7:::
lp:*:18226:0:99999:7:::
mail:*:18226:0:99999:7:::
news:*:18226:0:99999:7:::
uucp:*:18226:0:99999:7:::
proxy:*:18226:0:99999:7:::
www-data:*:18226:0:99999:7:::
backup:*:18226:0:99999:7:::
list:*:18226:0:99999:7:::
irc:*:18226:0:99999:7:::
gnats:*:18226:0:99999:7:::
nobody:*:18226:0:99999:7:::
_apt:*:18226:0:99999:7:::
systemd-timesync:*:18226:0:99999:7:::
systemd-network:*:18226:0:99999:7:::
systemd-resolve:*:18226:0:99999:7:::
Debian-exim:!:18226:0:99999:7:::
messagebus:*:18226:0:99999:7:::
debian-tor:*:18226:0:99999:7:::
dnsmasq:*:18226:0:99999:7:::
tss:*:18226:0:99999:7:::
geoclue:*:18226:0:99999:7:::
pulse:*:18226:0:99999:7:::
speech-dispatcher:!:18226:0:99999:7:::
usbmux:*:18226:0:99999:7:::
avahi:*:18226:0:99999:7:::
rtkit:*:18226:0:99999:7:::
saned:*:18226:0:99999:7:::
colord:*:18226:0:99999:7:::
Debian-gdm:*:18226:0:99999:7:::
`

func TestParseShadow(t *testing.T) {
	find := func(entries []ShadowEntry, username string) ShadowEntry {
		for _, e := range entries {
			if e.Username == username {
				return e
			}
		}
		t.Errorf("could not find entry: %q", username)
		return ShadowEntry{}
	}

	entries, err := ParseShadow(strings.NewReader(testShadow))
	if err != nil {
		t.Fatalf("ParseShadow() failed: %v", err)
	}

	if got, want := find(entries, "pulse"), (ShadowEntry{
		Username: "pulse",
		Password: ShadowPass{
			Mode: PassAccountDisabled,
		},
		LastChanged:       time.Date(2019, 11, 26, 0, 0, 0, 0, time.UTC),
		MaxChangeDays:     99999,
		WarnBeforeMaxDays: 7,
	}); !reflect.DeepEqual(got, want) {
		t.Errorf("entries['pulse'] = %v, want %v", got, want)
	}

	if got, want := find(entries, "www-data"), (ShadowEntry{
		Username: "www-data",
		Password: ShadowPass{
			Mode: PassAccountDisabled,
		},
		LastChanged:       time.Date(2019, 11, 26, 0, 0, 0, 0, time.UTC),
		MaxChangeDays:     99999,
		WarnBeforeMaxDays: 7,
	}); !reflect.DeepEqual(got, want) {
		t.Errorf("entries['www-data'] = %v, want %v", got, want)
	}
}

func TestShadowSerialize(t *testing.T) {
	shadow := strings.Replace(testShadow, "!", "*", -1)
	entries, err := ParseShadow(strings.NewReader(shadow))
	if err != nil {
		t.Fatal(err)
	}

	out, err := ShadowSerialize(entries)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal([]byte(shadow), out) {
		t.Errorf("shadow = %q, want %q", string(out), shadow)
	}
}
