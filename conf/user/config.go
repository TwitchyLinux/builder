package user

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/twitchylinux/builder/conf/kv"
)

// Config represents a static user/group configuration which can be mutated.
type Config struct {
	now      time.Time
	RootPath string
	users    []PasswdEntry
	groups   []GroupEntry

	shadow      []ShadowEntry
	shadowDirty bool

	createSkel []PasswdEntry
}

func decodeIDPair(conf map[string]string, lKey, rKey string, lDefault, rDefault int) (int, int, error) {
	var (
		lOut = lDefault
		rOut = rDefault
		err  error
	)

	if v, ok := conf[lKey]; ok {
		lOut, err = strconv.Atoi(v)
		if err != nil {
			return 0, 0, fmt.Errorf("parsing %q as int: %v", lKey, err)
		}
	}
	if v, ok := conf[rKey]; ok {
		rOut, err = strconv.Atoi(v)
		if err != nil {
			return 0, 0, fmt.Errorf("parsing %q as int: %v", rKey, err)
		}
	}

	return lOut, rOut, nil
}

func (m *Config) readAdduserConf(adduserPath string) (map[string]string, error) {
	f, err := os.Open(adduserPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return kv.ParseBasic(f)
}

func (m *Config) defaultHomedir(username string) (string, error) {
	conf, err := m.readAdduserConf(filepath.Join(m.RootPath, "etc", "adduser.conf"))
	if err != nil {
		if os.IsNotExist(err) {
			return "/home/" + username, nil
		}
		return "", err
	}
	if home, ok := conf["DHOME"]; ok {
		return filepath.Join(home, username), nil
	}
	return "/home/" + username, nil
}

func (m *Config) defaultShell() (string, error) {
	conf, err := m.readAdduserConf(filepath.Join(m.RootPath, "etc", "adduser.conf"))
	if err != nil {
		if os.IsNotExist(err) {
			return "/bin/bash", nil
		}
		return "", err
	}
	if shell, ok := conf["DSHELL"]; ok {
		return shell, nil
	}
	return "/bin/bash", nil
}

func (m *Config) systemUIDRange() (int, int, error) {
	conf, err := m.readAdduserConf(filepath.Join(m.RootPath, "etc", "adduser.conf"))
	if err != nil {
		if os.IsNotExist(err) {
			return 100, 999, nil // use reasonable defaults.
		}
		return 0, 0, err
	}

	return decodeIDPair(conf, "FIRST_SYSTEM_UID", "LAST_SYSTEM_UID", 100, 999)
}

func (m *Config) userUIDRange() (int, int, error) {
	conf, err := m.readAdduserConf(filepath.Join(m.RootPath, "etc", "adduser.conf"))
	if err != nil {
		if os.IsNotExist(err) {
			return 1000, 59999, nil // use reasonable defaults.
		}
		return 0, 0, err
	}

	return decodeIDPair(conf, "FIRST_UID", "LAST_UID", 1000, 59999)
}

func (m *Config) systemGIDRange() (int, int, error) {
	conf, err := m.readAdduserConf(filepath.Join(m.RootPath, "etc", "adduser.conf"))
	if err != nil {
		if os.IsNotExist(err) {
			return 100, 999, nil // use reasonable defaults.
		}
		return 0, 0, err
	}

	return decodeIDPair(conf, "FIRST_SYSTEM_GID", "LAST_SYSTEM_GID", 100, 999)
}

func (m *Config) userGIDRange() (int, int, error) {
	conf, err := m.readAdduserConf(filepath.Join(m.RootPath, "etc", "adduser.conf"))
	if err != nil {
		if os.IsNotExist(err) {
			return 1000, 59999, nil // use reasonable defaults.
		}
		return 0, 0, err
	}

	return decodeIDPair(conf, "FIRST_GID", "LAST_GID", 1000, 59999)
}

// SetPassword sets the password on the account.
func (m *Config) SetPassword(name string, pass ShadowPass) error {
	idx := -1
	for i, u := range m.shadow {
		if u.Username == name {
			idx = i
			break
		}
	}

	if idx > -1 {
		m.shadow[idx].Password = pass
	} else {
		m.shadow = append(m.shadow, ShadowEntry{
			Username:          name,
			LastChanged:       m.now,
			MaxChangeDays:     99999,
			WarnBeforeMaxDays: 7,
		})
	}
	m.shadowDirty = true
	return nil
}

// OptSystemAccount indicates any created user should use the system UID/GID
// ranges.
func OptSystemAccount() AddUserOpt {
	return AddUserOpt{sysAccount: true}
}

// OptNoUserGroup indicates no user group should be created for any
// created user.
func OptNoUserGroup() AddUserOpt {
	return AddUserOpt{noUserGroup: true}
}

// OptCreateSkel indicates that a skeleton homedir should be created
// based on /etc/skel.
func OptCreateSkel() AddUserOpt {
	return AddUserOpt{skel: true}
}

// OptErrIfExists will cause UpsertUser to fail if the user already exists.
func OptErrIfExists() AddUserOpt {
	return AddUserOpt{errIfExists: true}
}

// OptHomedir configures the homedir for the user.
func OptHomedir(dir string) AddUserOpt {
	return AddUserOpt{homeDir: dir}
}

// OptShell configures the shell for the user.
func OptShell(shell string) AddUserOpt {
	return AddUserOpt{shell: shell}
}

// AddUserOpt represents an option configuring an invocation to UpsertUser().
type AddUserOpt struct {
	sysAccount  bool
	noUserGroup bool
	skel        bool
	errIfExists bool
	homeDir     string
	shell       string
}

func flattenAddUserOpts(opts ...AddUserOpt) AddUserOpt {
	out := AddUserOpt{}
	for _, o := range opts {
		if o.sysAccount {
			out.sysAccount = true
		}
		if o.errIfExists {
			out.errIfExists = true
		}
		if o.homeDir != "" {
			out.homeDir = o.homeDir
		}
		if o.shell != "" {
			out.shell = o.shell
		}
		if o.noUserGroup {
			out.noUserGroup = true
		}
		if o.skel {
			out.skel = true
		}
	}
	return out
}

func (m *Config) userIndex(name string) int {
	idx := -1
	for i, u := range m.users {
		if u.Username == name {
			idx = i
			break
		}
	}
	return idx
}
func (m *Config) groupIndex(name string) int {
	idx := -1
	for i, u := range m.groups {
		if u.Name == name {
			idx = i
			break
		}
	}
	return idx
}

// UpsertMembership adds the named user to the named group, if not already
// a member.
func (m *Config) UpsertMembership(user, group string) error {
	if m.userIndex(user) < 0 {
		return fmt.Errorf("user %q does not exist", user)
	}
	idx := m.groupIndex(group)
	if idx < 0 {
		return fmt.Errorf("group %q does not exist", group)
	}

	// Check if the user is already a member.
	for _, m := range m.groups[idx].Users {
		if m == user {
			return nil // Already a member.
		}
	}

	// Otherwise, add the membership.
	m.groups[idx].Users = append(m.groups[idx].Users, user)
	return nil
}

// UpsertUser creates the given user if it does not exist, or updates the
// homedir, shell, and user info otherwise.
func (m *Config) UpsertUser(name string, opts ...AddUserOpt) error {
	options := flattenAddUserOpts(opts...)
	var userInfo = ""

	idx := m.userIndex(name)
	if options.errIfExists && idx >= 0 {
		return os.ErrExist
	}

	if idx >= 0 { // user exists, update values
		if options.homeDir != "" {
			m.users[idx].HomeDir = options.homeDir
		}
		if options.shell != "" {
			m.users[idx].ShellPath = options.shell
		}
		if userInfo != "" {
			m.users[idx].UserInfo = userInfo
		}
		return nil
	}

	// User doesnt exist, determine defaults.
	usr, err := m.newUserEntry(name, options.sysAccount, options.homeDir, options.shell, userInfo)
	if err != nil {
		return fmt.Errorf("generating user: %v", err)
	}
	if !options.noUserGroup {
		m.groups = append(m.groups, GroupEntry{ID: usr.GID, Name: usr.Username, Pass: "x"})
	}
	m.users = append(m.users, usr)
	if options.skel {
		m.createSkel = append(m.createSkel, usr)
	}
	return m.SetPassword(name, ShadowPass{Mode: PassAccountDisabled})
}

func (m *Config) newUserEntry(name string, systemUser bool, homeDir, shell, userInfo string) (PasswdEntry, error) {
	var (
		usr                    = PasswdEntry{Username: name, Password: PasswdPass{Mode: PassShadow}, UserInfo: userInfo}
		lowerUID, upperUID int = 0, 0
		lowerGID, upperGID int = 0, 0
		err                error
	)
	if systemUser {
		if lowerUID, upperUID, err = m.systemUIDRange(); err != nil {
			return PasswdEntry{}, err
		}
		if lowerGID, upperGID, err = m.systemGIDRange(); err != nil {
			return PasswdEntry{}, err
		}
	} else {
		if lowerUID, upperUID, err = m.userUIDRange(); err != nil {
			return PasswdEntry{}, err
		}
		if lowerGID, upperGID, err = m.userGIDRange(); err != nil {
			return PasswdEntry{}, err
		}
	}

	if usr.UID, err = m.findFreeUIDInRange(lowerUID, upperUID); err != nil {
		return PasswdEntry{}, err
	}
	if usr.GID, err = m.findFreeGIDInRange(lowerGID, upperGID, usr.UID); err != nil {
		return PasswdEntry{}, err
	}

	if homeDir == "" {
		if usr.HomeDir, err = m.defaultHomedir(usr.Username); err != nil {
			return PasswdEntry{}, err
		}
	} else {
		usr.HomeDir = homeDir
	}
	if shell == "" {
		if usr.ShellPath, err = m.defaultShell(); err != nil {
			return PasswdEntry{}, err
		}
	} else {
		usr.ShellPath = shell
	}
	return usr, nil
}

func (m *Config) findFreeUIDInRange(lower, upper int) (int, error) {
	seenUIDs := map[int]struct{}{}
	for _, u := range m.users {
		seenUIDs[u.UID] = struct{}{}
	}
	for i := lower; i < upper; i++ {
		if _, used := seenUIDs[i]; !used {
			return i, nil
		}
	}
	return 0, fmt.Errorf("no free UIDs from %d to %d", lower, upper)
}

func (m *Config) findFreeGIDInRange(lower, upper, preferred int) (int, error) {
	seenGIDs := map[int]struct{}{}
	for _, u := range m.groups {
		seenGIDs[u.ID] = struct{}{}
	}
	if _, used := seenGIDs[preferred]; !used {
		return preferred, nil
	}
	for i := lower; i < upper; i++ {
		if _, used := seenGIDs[i]; !used {
			return i, nil
		}
	}
	return 0, fmt.Errorf("no free GIDs from %d to %d", lower, upper)
}

// Flush writes the config, including any modifications, to the filesystem.
func (m *Config) Flush() error {
	passwd, err := PasswdSerialize(m.users)
	if err != nil {
		return err
	}
	groups, err := GroupSerialize(m.groups)
	if err != nil {
		return err
	}
	shadow, err := ShadowSerialize(m.shadow)
	if err != nil {
		return err
	}

	if err := writeFile(filepath.Join(m.RootPath, "etc", "passwd"), passwd, 0644); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(m.RootPath, "etc", "group"), groups, 0644); err != nil {
		return err
	}
	if m.shadowDirty {
		if err := writeFile(filepath.Join(m.RootPath, "etc", "shadow"), shadow, 0640); err != nil {
			return err
		}
	}
	return m.flushSkel()
}

func (c *Config) flushSkel() error {
	for _, u := range c.createSkel {
		if err := c.makeHomedir(u.HomeDir, "/etc/skel", u.UID, u.GID); err != nil {
			return fmt.Errorf("failed to create home directory for %s: %v", u.Username, err)
		}
	}
	c.createSkel = nil
	return nil
}

func (c *Config) makeHomedir(home, skel string, uid, gid int) error {
	home = filepath.Join(c.RootPath, home)

	if err := os.Mkdir(home, 02700); err != nil {
		return fmt.Errorf("mkdir failed: %v", err)
	}

	cmd := exec.Command("cp", "-a", filepath.Clean(skel)+"/.", home)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed copying %q to %q: %v", skel, home, err)
	}
	cmd = exec.Command("chown", "-R", fmt.Sprintf("%d:%d", uid, gid), home)
	return cmd.Run()
}

// ReadConfig reads the static user/group configuration in the
// given filesystem root.
func ReadConfig(rootPath string) (*Config, error) {
	m := Config{now: time.Now().UTC(), RootPath: rootPath}
	var err error

	if m.users, err = readUsers(rootPath); err != nil {
		return nil, fmt.Errorf("reading users: %v", err)
	}
	if m.groups, err = readGroups(rootPath); err != nil {
		return nil, fmt.Errorf("reading groups: %v", err)
	}
	if m.shadow, err = readShadow(rootPath); err != nil {
		return nil, fmt.Errorf("reading shadow: %v", err)
	}
	return &m, nil
}

func writeFile(path string, data []byte, perm os.FileMode) error {
	var uid, gid int

	if s, err := os.Stat(path); err == nil {
		perm = s.Mode()
		uid = int(s.Sys().(*syscall.Stat_t).Uid)
		gid = int(s.Sys().(*syscall.Stat_t).Gid)
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, bytes.NewReader(data)); err != nil {
		return err
	}
	if err := f.Chown(uid, gid); err != nil {
		return err
	}
	return f.Close()
}

func readGroups(rootPath string) ([]GroupEntry, error) {
	g, err := os.Open(filepath.Join(rootPath, "etc", "group"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer g.Close()
	return ParseGroup(g)
}

func readUsers(rootPath string) ([]PasswdEntry, error) {
	p, err := os.Open(filepath.Join(rootPath, "etc", "passwd"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer p.Close()
	return ParsePasswd(p)
}

func readShadow(rootPath string) ([]ShadowEntry, error) {
	p, err := os.Open(filepath.Join(rootPath, "etc", "shadow"))
	if err != nil {
		if os.IsNotExist(err) || os.IsPermission(err) {
			return nil, nil
		}
		return nil, err
	}
	defer p.Close()
	return ParseShadow(p)
}
