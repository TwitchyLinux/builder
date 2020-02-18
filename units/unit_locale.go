package units

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Locale is a unit which sets the system locale and timezone.
type Locale struct {
	Area, Zone string
	Generate   []string
	Default    string
}

// Name implements Unit.
func (d *Locale) Name() string {
	return "Locale"
}

func (d *Locale) writeTZ(ctx context.Context, opts *Opts, chroot *Chroot) error {
	if err := ioutil.WriteFile(filepath.Join(opts.Dir, "tz-data"), []byte(`
tzdata tzdata/Areas select `+d.Area+`
	tzdata tzdata/Zones/Europe select `+d.Zone+`
`), 0644); err != nil {
		return err
	}
	defer os.Remove(filepath.Join(opts.Dir, "tz-data"))

	cmd, err := chroot.CmdContext(ctx, opts, "debconf-set-selections", "/tz-data")
	if err != nil {
		return err
	}
	cmd.Env = localeEnv
	cmd.Stdout = opts.L.Stdout()
	cmd.Stderr = opts.L.Stderr()
	return cmd.Run()
}

func (d *Locale) updateLocaleGen(ctx context.Context, opts *Opts) error {
	var out bytes.Buffer
	genConf, err := ioutil.ReadFile(filepath.Join(opts.Dir, "etc", "locale.gen"))
	if err != nil {
		return err
	}

lineLoop:
	for _, line := range strings.Split(string(genConf), "\n") {
		for _, gen := range d.Generate {
			if strings.HasPrefix(line, "# ") && strings.Contains(line, gen) {
				out.WriteString(line[2:])
				out.WriteRune('\n')
				continue lineLoop
			}
		}
		out.WriteString(line)
		out.WriteRune('\n')
	}

	return ioutil.WriteFile(filepath.Join(opts.Dir, "etc", "locale.gen"), out.Bytes(), 0644)
}

// Run implements Unit.
func (d *Locale) Run(ctx context.Context, opts Opts) error {
	chroot, err := prepareChroot(opts.Dir)
	if err != nil {
		return err
	}
	defer chroot.Close()

	if err := d.writeTZ(ctx, &opts, chroot); err != nil {
		return err
	}

	if err := chroot.AptInstall(ctx, &opts, "locales"); err != nil {
		return err
	}

	if err := d.updateLocaleGen(ctx, &opts); err != nil {
		return err
	}

	cmd, err := chroot.CmdContext(ctx, &opts, "locale-gen", d.Default)
	if err != nil {
		return err
	}
	cmd.Env = localeEnv
	cmd.Stdout = opts.L.Stdout()
	cmd.Stderr = opts.L.Stderr()
	if err := cmd.Run(); err != nil {
		return err
	}

	if cmd, err = chroot.CmdContext(ctx, &opts, "debconf-set-selections"); err != nil {
		return err
	}
	cmd.Env = localeEnv
	cmd.Stdin = strings.NewReader("locales locales/locales_to_be_generated multiselect " + strings.Join(d.Generate, ", ") + "\n")
	cmd.Stdout = opts.L.Stdout()
	cmd.Stderr = opts.L.Stderr()
	if err := cmd.Run(); err != nil {
		return err
	}

	if cmd, err = chroot.CmdContext(ctx, &opts, "debconf-set-selections"); err != nil {
		return err
	}
	cmd.Env = localeEnv
	cmd.Stdin = strings.NewReader("locales locales/default_environment_locale select " + d.Default + "\n")
	cmd.Stdout = opts.L.Stdout()
	cmd.Stderr = opts.L.Stderr()
	if err := cmd.Run(); err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(opts.Dir, "etc", "timezone"), []byte(d.Area+"/"+d.Zone+"\n"), 0644); err != nil {
		return err
	}

	if cmd, err = chroot.CmdContext(ctx, &opts, "dpkg-reconfigure", "--frontend=noninteractive", "locales"); err != nil {
		return err
	}
	cmd.Env = localeEnv
	cmd.Stdout = opts.L.Stdout()
	cmd.Stderr = opts.L.Stderr()
	if err := cmd.Run(); err != nil {
		return err
	}

	os.Remove(filepath.Join(opts.Dir, "etc", "localtime"))
	return os.Symlink("../usr/share/zoneinfo/"+d.Area+"/"+d.Zone, filepath.Join(opts.Dir, "etc", "localtime"))
}
