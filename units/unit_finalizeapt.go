package units

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// FinalizeApt configures apt into an ideal state.
type FinalizeApt struct {
	Track string
}

// Name implements Unit.
func (u *FinalizeApt) Name() string {
	return "Finalize-apt"
}

func (u *FinalizeApt) fixAptSources(path string) error {
	d, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	uncommentOnly := strings.Contains(string(d), "deb-src")
	var out strings.Builder
	for _, line := range strings.Split(string(d), "\n") {
		out.WriteString(line)
		out.WriteRune('\n')

		spl := strings.Split(line, " ")
		if !uncommentOnly && len(spl) > 1 && spl[0] == "deb" {
			out.WriteString("deb-src " + strings.Join(spl[1:], " "))
			out.WriteRune('\n')
		} else if uncommentOnly && len(spl) > 2 && spl[0] == "#" && spl[1] == "deb-src" {
			out.WriteString(strings.Join(spl[1:], " "))
			out.WriteRune('\n')
		}
	}

	s := out.String()

	// Hack in the non-free component.
	if !strings.Contains(s, "non-free contrib") {
		s = strings.Replace(s, " "+u.Track+" main\n", " "+u.Track+" main non-free contrib\n", -1)
	}
	return ioutil.WriteFile(path, []byte(s), 0644)
}

// Run implements Unit.
func (u *FinalizeApt) Run(ctx context.Context, opts Opts) error {
	if err := u.fixAptSources(filepath.Join(opts.Dir, "etc", "apt", "sources.list")); err != nil {
		return err
	}

	if opts.DebProxy != "" {
		if err := ioutil.WriteFile(filepath.Join(opts.Dir, "etc", "apt", "apt.conf.d", "05-temp-install-proxy"),
			[]byte("Acquire::http::proxy::deb.debian.org \"http://"+opts.DebProxy+"/\";\n"), 0644); err != nil {
			return err
		}
	}

	chroot, err := prepareChroot(opts.Dir)
	if err != nil {
		return err
	}
	defer chroot.Close()

	// TODO: Detect if host ufw is enabled and would block traffic.
	cmd, err := chroot.CmdContext(ctx, &opts, "apt-get", "--fix-broken", "-y", "install")
	if err != nil {
		return err
	}
	cmd.Stdout = opts.L.Stdout()
	cmd.Stderr = opts.L.Stderr()
	if opts.DebProxy != "" {
		cmd.Env = append(cmd.Env, "http_proxy=http://"+opts.DebProxy)
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd, err = chroot.CmdContext(ctx, &opts, "apt-get", "update")
	if err != nil {
		return err
	}
	cmd.Stdout = opts.L.Stdout()
	cmd.Stderr = opts.L.Stderr()
	if opts.DebProxy != "" {
		cmd.Env = append(cmd.Env, "http_proxy=http://"+opts.DebProxy)
	}
	return cmd.Run()
}
