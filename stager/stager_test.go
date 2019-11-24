package stager

import (
	"reflect"
	"testing"

	"github.com/twitchylinux/builder/units"
)

func getUnit(t *testing.T, units []units.Unit, typ reflect.Type) units.Unit {
	t.Helper()
	for i := range units {
		if reflect.TypeOf(units[i]) == typ {
			return units[i]
		}
	}

	t.Fatalf("Expected unit of type %v, found none", typ)
	panic("unreachable")
}

func getUnits(t *testing.T, in []units.Unit, typ reflect.Type) []units.Unit {
	t.Helper()
	var out []units.Unit
	for i := range in {
		if reflect.TypeOf(in[i]) == typ {
			out = append(out, in[i])
		}
	}
	return out
}

func TestLoadLocale(t *testing.T) {
	c, err := UnitsFromConfig("testdata/locale")
	if err != nil {
		t.Fatal(err)
	}

	l := getUnit(t, c, reflect.TypeOf(&units.Locale{})).(*units.Locale)
	if got, want := l, (&units.Locale{
		Area:     "HNNNNNG",
		Zone:     "Los_Angeles",
		Generate: []string{"en_US.UTF-8 UTF-8", "en_US ISO-8859-1"},
		Default:  "en_US.UTF-8",
	}); !reflect.DeepEqual(got, want) {
		t.Errorf("locale = %v, want %v", got, want)
	}
}

func TestLoadDebootstrapDefaults(t *testing.T) {
	c, err := UnitsFromConfig("testdata/empty")
	if err != nil {
		t.Fatal(err)
	}

	got := getUnit(t, c, reflect.TypeOf(&units.Debootstrap{})).(*units.Debootstrap)
	if want := (&units.Debootstrap{
		Track: debootstrapDefault.Track,
		URL:   debootstrapDefault.URL,
	}); !reflect.DeepEqual(got, want) {
		t.Errorf("debian = %v, want %v", got, want)
	}
}

func TestLoadGolangDefaults(t *testing.T) {
	c, err := UnitsFromConfig("testdata/empty")
	if err != nil {
		t.Fatal(err)
	}

	got := getUnit(t, c, reflect.TypeOf(&units.Golang{})).(*units.Golang)
	if want := (&units.Golang{
		Version: golangDefault.Version,
		URL:     golangDefault.URL,
		SHA256:  golangDefault.SHA256,
	}); !reflect.DeepEqual(got, want) {
		t.Errorf("golang = %v, want %v", got, want)
	}
}

func TestLoadLinuxDefaults(t *testing.T) {
	c, err := UnitsFromConfig("testdata/empty")
	if err != nil {
		t.Fatal(err)
	}

	got := getUnit(t, c, reflect.TypeOf(&units.Linux{})).(*units.Linux)
	if want := (&units.Linux{
		Version: linuxDefault.Version,
		URL:     linuxDefault.URL,
		SHA256:  linuxDefault.SHA256,
	}); !reflect.DeepEqual(got, want) {
		t.Errorf("linux = %v, want %v", got, want)
	}
}

func TestLoadGraphical(t *testing.T) {
	c, err := UnitsFromConfig("testdata/graphics")
	if err != nil {
		t.Fatal(err)
	}

	gnome := getUnit(t, c, reflect.TypeOf(&units.Gnome{})).(*units.Gnome)
	if got, want := gnome.NeedPkgs, []string{"test", "yolo"}; !reflect.DeepEqual(got, want) {
		t.Errorf("gnome.NeedPkgs = %v, want %v", got, want)
	}
}

func TestLoadGraphicalDefaults(t *testing.T) {
	c, err := UnitsFromConfig("testdata/empty")
	if err != nil {
		t.Fatal(err)
	}

	gnome := getUnit(t, c, reflect.TypeOf(&units.Gnome{})).(*units.Gnome)
	if got, want := gnome.NeedPkgs, graphicalEnvDefault.Packages; !reflect.DeepEqual(got, want) {
		t.Errorf("gnome.NeedPkgs = %v, want %v", got, want)
	}
}

func TestLoadGraphicalDisabled(t *testing.T) {
	c, err := UnitsFromConfig("testdata/graphics_none")
	if err != nil {
		t.Fatal(err)
	}

	gnome := getUnit(t, c, reflect.TypeOf(&units.Gnome{})).(*units.Gnome)
	if gnome != nil {
		t.Errorf("expected nil unit, got %v", gnome)
	}
}

func TestLoadPostInstall(t *testing.T) {
	c, err := UnitsFromConfig("testdata/post_base_install")
	if err != nil {
		t.Fatal(err)
	}

	tools := getUnits(t, c, reflect.TypeOf(&units.InstallTools{}))

	if got, want := tools, []units.Unit{
		&units.InstallTools{
			UnitName: "cli",
			Pkgs:     []string{"screen", "htop"},
			Order:    5,
		},
		&units.InstallTools{
			UnitName: "med",
			Pkgs:     []string{"med"},
			Order:    2,
		},
		&units.InstallTools{
			UnitName: "last",
			Pkgs:     []string{"last"},
			Order:    1,
		},
	}; !reflect.DeepEqual(got, want) {
		for i := range want {
			t.Errorf("tools[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestLoadComposites(t *testing.T) {
	c, err := UnitsFromConfig("testdata/composite")
	if err != nil {
		t.Fatal(err)
	}

	i := getUnit(t, c, reflect.TypeOf(&units.Composite{})).(*units.Composite)
	if got, want := i.UnitName, "cli"; got != want {
		t.Errorf("UnitName = %v, want %v", got, want)
	}
	if got, want := len(i.Ops), 3; got != want {
		t.Fatalf("len(Ops) = %v, want %v", got, want)
	}

	if got, want := i.Ops[0], (&units.InstallTools{
		UnitName: "cli",
		Pkgs:     []string{"screen", "htop"},
	}); !reflect.DeepEqual(got, want) {
		t.Errorf("Op[0] = %v, want %v", got, want)
	}
	if got, want := i.Ops[1], (&units.Download{
		URL: "https://dl.google.com/linux/linux_signing_key.pub",
		To:  "/chrome-signing-key.pub",
	}); !reflect.DeepEqual(got, want) {
		t.Errorf("Op[1] = %v, want %v", got, want)
	}
	if got, want := i.Ops[2], (&units.Cmd{
		Bin:  "apt-key",
		Args: []string{"add", "/chrome-signing-key.pub"},
	}); !reflect.DeepEqual(got, want) {
		t.Errorf("Op[2] = %v, want %v", got, want)
	}
}

func TestLoadStageConf(t *testing.T) {
	c, err := UnitsFromConfig("../resources/stage-conf")
	if err != nil {
		t.Fatal(err)
	}

toolLoop:
	for _, tool := range []string{"fs-tools", "cli-tools", "wifi", "compression-tools", "gui-dev-tools"} {
		for _, u := range c {
			if inst, ok := u.(*units.InstallTools); ok && inst.UnitName == tool {
				continue toolLoop
			}
		}
		t.Errorf("Could not find stage for install.post_base.%s", tool)
	}

	linux := getUnit(t, c, reflect.TypeOf(&units.Linux{})).(*units.Linux)
	if len(linux.BuildDepPkgs) == 0 {
		t.Error("len(linux.BuildDepPkgs) = 0, want >0")
	}
}

func TestStageConfOrdering(t *testing.T) {
	c, err := UnitsFromConfig("../resources/stage-conf")
	if err != nil {
		t.Fatal(err)
	}

	stageFinder := func(typ reflect.Type, unitName string) func() int {
		return func() int {
			for i := range c {
				if unitName != "" {
					if install, ok := c[i].(*units.InstallTools); ok && install.UnitName == unitName {
						return i
					}
				} else if reflect.TypeOf(c[i]) == typ {
					return i
				}
			}
			return 999
		}
	}

	tcs := []struct {
		name   string
		before func() int
		after  func() int
	}{
		{
			name:   "Preflight first",
			before: stageFinder(reflect.TypeOf(&units.Preflight{}), ""),
			after:  stageFinder(reflect.TypeOf(&units.Debootstrap{}), ""),
		},
		{
			name:   "Apt before linux",
			before: stageFinder(reflect.TypeOf(&units.FinalizeApt{}), ""),
			after:  stageFinder(reflect.TypeOf(&units.Linux{}), ""),
		},
		{
			name:   "Locale before linux",
			before: stageFinder(reflect.TypeOf(&units.Locale{}), ""),
			after:  stageFinder(reflect.TypeOf(&units.Linux{}), ""),
		},
		{
			name:   "Linux before customization",
			before: stageFinder(reflect.TypeOf(&units.Linux{}), ""),
			after:  stageFinder(reflect.TypeOf(&units.ShellCustomization{}), ""),
		},
		{
			name:   "FS tools before Gnome",
			before: stageFinder(reflect.TypeOf(&units.InstallTools{}), "fs-tools"),
			after:  stageFinder(reflect.TypeOf(&units.Gnome{}), ""),
		},
		{
			name:   "FS tools before firmware",
			before: stageFinder(reflect.TypeOf(&units.InstallTools{}), "fs-tools"),
			after:  stageFinder(reflect.TypeOf(&units.InstallTools{}), "firmware"),
		},
		{
			name:   "GUI tools after Gnome",
			before: stageFinder(reflect.TypeOf(&units.Gnome{}), ""),
			after:  stageFinder(reflect.TypeOf(&units.InstallTools{}), "gui-dev-tools"),
		},
		{
			name:   "Clean before Grub",
			before: stageFinder(reflect.TypeOf(&units.Clean{}), ""),
			after:  stageFinder(reflect.TypeOf(&units.Grub2{}), ""),
		},
		{
			name:   "Grub last",
			before: func() int { return len(c) - 1 },
			after:  stageFinder(reflect.TypeOf(&units.Grub2{}), ""),
		},
	}

	for _, i := range c {
		t.Logf("%+v\n", i)
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			before, after := tc.before(), tc.after()
			if before == 999 {
				t.Fatal("could not find earlier stage")
			}
			if after == 999 {
				t.Fatal("could not find later stage")
			}
			if before > after {
				t.Errorf("%T (%+v) was before %T (%+v)", c[before], c[before], c[after], c[after])
			}
		})
	}
}

func TestLoadUnionsOverlaps(t *testing.T) {
	c, err := UnitsFromConfig("testdata/overlap_across_files")
	if err != nil {
		t.Fatal(err)
	}

	expect := []*units.InstallTools{
		{
			UnitName: "cli",
			Pkgs:     []string{"screen", "htop"},
			Order:    5,
		},
		{
			UnitName: "med",
			Pkgs:     []string{"med"},
			Order:    2,
		},
	}

expectLoop:
	for _, exp := range expect {
		for _, unit := range c {
			it, ok := unit.(*units.InstallTools)
			if !ok {
				continue
			}
			if reflect.DeepEqual(it, exp) {
				continue expectLoop
			}
		}
		t.Errorf("missing unit: %+v", exp)
	}
}
