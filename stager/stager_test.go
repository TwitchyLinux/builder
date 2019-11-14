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
