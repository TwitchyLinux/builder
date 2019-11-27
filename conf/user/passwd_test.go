package user

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

var testPasswd = `root:x:0:0:root:/root:/bin/bash
daemon:x:1:1:daemon:/usr/sbin:/usr/sbin/nologin
bin:x:2:2:bin:/bin:/usr/sbin/nologin
sys:x:3:3:sys:/dev:/usr/sbin/nologin
sync:x:4:65534:sync:/bin:/bin/sync
games:x:5:60:games:/usr/games:/usr/sbin/nologin
man:x:6:12:man:/var/cache/man:/usr/sbin/nologin
lp:x:7:7:lp:/var/spool/lpd:/usr/sbin/nologin
mail:x:8:8:mail:/var/mail:/usr/sbin/nologin
news:x:9:9:news:/var/spool/news:/usr/sbin/nologin
uucp:x:10:10:uucp:/var/spool/uucp:/usr/sbin/nologin
proxy:x:13:13:proxy:/bin:/usr/sbin/nologin
www-data:x:33:33:www-data:/var/www:/usr/sbin/nologin
backup:x:34:34:backup:/var/backups:/usr/sbin/nologin
list:x:38:38:Mailing List Manager:/var/list:/usr/sbin/nologin
irc:x:39:39:ircd:/var/run/ircd:/usr/sbin/nologin
gnats:x:41:41:Gnats Bug-Reporting System (admin):/var/lib/gnats:/usr/sbin/nologin
nobody:x:65534:65534:nobody:/nonexistent:/usr/sbin/nologin
_apt:x:100:65534::/nonexistent:/usr/sbin/nologin
systemd-timesync:x:101:102:systemd Time Synchronization,,,:/run/systemd:/usr/sbin/nologin
systemd-network:x:102:103:systemd Network Management,,,:/run/systemd:/usr/sbin/nologin
systemd-resolve:x:103:104:systemd Resolver,,,:/run/systemd:/usr/sbin/nologin
Debian-exim:x:104:110::/var/spool/exim4:/usr/sbin/nologin
twl:x:1000:1000:,,,:/home/twl:/bin/bash
messagebus:x:105:112::/nonexistent:/usr/sbin/nologin
debian-tor:x:106:114::/var/lib/tor:/bin/false
dnsmasq:x:107:65534:dnsmasq,,,:/var/lib/misc:/usr/sbin/nologin
tss:x:108:116:TPM2 software stack,,,:/var/lib/tpm:/bin/false
geoclue:x:109:117::/var/lib/geoclue:/usr/sbin/nologin
pulse:x:110:118:PulseAudio daemon,,,:/var/run/pulse:/usr/sbin/nologin
speech-dispatcher:x:111:29:Speech Dispatcher,,,:/var/run/speech-dispatcher:/bin/false
usbmux:x:112:46:usbmux daemon,,,:/var/lib/usbmux:/usr/sbin/nologin
avahi:x:113:120:Avahi mDNS daemon,,,:/var/run/avahi-daemon:/usr/sbin/nologin
rtkit:x:114:121:RealtimeKit,,,:/proc:/usr/sbin/nologin
saned:x:115:123::/var/lib/saned:/usr/sbin/nologin
colord:x:116:124:colord colour management daemon,,,:/var/lib/colord:/usr/sbin/nologin
Debian-gdm:x:117:125:Gnome Display Manager:/var/lib/gdm3:/bin/false
`

func TestParsePasswd(t *testing.T) {
	find := func(entries []PasswdEntry, username string) PasswdEntry {
		for _, e := range entries {
			if e.Username == username {
				return e
			}
		}
		t.Errorf("could not find entry: %q", username)
		return PasswdEntry{}
	}

	entries, err := ParsePasswd(strings.NewReader(testPasswd))
	if err != nil {
		t.Fatalf("ParsePasswd() failed: %v", err)
	}

	if got, want := find(entries, "colord"), (PasswdEntry{
		Username: "colord",
		Password: PasswdPass{
			Mode: PassShadow,
		},
		UID:       116,
		GID:       124,
		UserInfo:  "colord colour management daemon,,,",
		HomeDir:   "/var/lib/colord",
		ShellPath: "/usr/sbin/nologin",
	}); !reflect.DeepEqual(got, want) {
		t.Errorf("entries['colord'] = %v, want %v", got, want)
	}
}

func TestPasswdSerialize(t *testing.T) {
	shadow := strings.Replace(testPasswd, "!", "*", -1)
	entries, err := ParsePasswd(strings.NewReader(shadow))
	if err != nil {
		t.Fatal(err)
	}

	out, err := PasswdSerialize(entries)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal([]byte(shadow), out) {
		t.Errorf("passwd = %q, want %q", string(out), shadow)
	}
}
