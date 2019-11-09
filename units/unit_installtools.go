package units

import (
	"context"
)

var (
	fsToolsInstall = &InstallTools{
		name: "fs-tools",
		pkgs: []string{
			"e2fsprogs", "e2fsck-static", "gpart", "parted", "kpartx",
			"libfuse2", "ntfs-3g", "fuse2fs", "fuseiso", "fatcat", "fatattr", "exfat-utils", "exfat-fuse",
			"cpio", "disktype", "cpio-doc", "multipath-tools", "smartmontools", "dosfstools",
			"sg3-utils", "squashfs-tools", "libext2fs2", "apt-file",
		},
	}
	netToolsInstall = &InstallTools{
		name: "net-tools",
		pkgs: []string{
			"curl", "curlftpfs", "net-tools", "netbase", "netcat-openbsd", "netwox",
			"iproute2", "iproute2-doc", "tcpdump", "tcpstat", "traceroute",
			"inetutils-ftp", "inetutils-ping", "inetutils-tools",
			"arping", "arptables", "dnsutils", "irssi",
		},
	}
	compressionToolsInstall = &InstallTools{
		name: "compression-tools",
		pkgs: []string{"lzip", "unzip", "zip", "tar", "rar", "unrar"},
	}
	cliToolsInstall = &InstallTools{
		name: "cli-tools",
		pkgs: []string{
			"htop", "screen", "tmux",
			"nano", "vim", "vim-doc",
			"git", "git-doc", "subversion", "subversion-tools",
			"lua5.2", "python2.7", "python-pip", "python-numpy", "python-dev", "python-numpy-doc", "python3", "python3-numpy",
			"figlet",
			"sharutils", "sharutils-doc",
			"pm-utils", "pciutils", "sysstat", "dpkg-dev",
			"gnupg", "jq", "tidy", "dstat", "info", "attr",
		},
	}
	nmapToolsInstall = &InstallTools{name: "nmap-tools", pkgs: []string{"nmap", "ncat", "ndiff", "zenmap"}}
	usbInstall       = &InstallTools{
		name: "usb",
		pkgs: []string{"usbutils", "libusb-1.0-0", "libusb-1.0-0-dev", "libusb-1.0-doc"},
	}
	cToolchainInstall = &InstallTools{
		name: "c-toolchain",
		pkgs: []string{"build-essential", "cmake", "sqlite3", "libsqlite3-0", "libsqlite3-dev", "sqlite3-doc"},
	}
	wifiInstall = &InstallTools{
		name: "wifi",
		pkgs: []string{"iw", "wireless-tools", "wpasupplicant", "rfkill", "net-tools"},
	}
)

// InstallTools is a unit which installs packages from apt.
type InstallTools struct {
	name string
	pkgs []string
}

// Name implements Unit.
func (i *InstallTools) Name() string {
	return i.name
}

// Run implements Unit.
func (i *InstallTools) Run(ctx context.Context, opts Opts) error {
	chroot, err := prepareChroot(opts.Dir)
	if err != nil {
		return err
	}
	defer chroot.Close()

	return chroot.AptInstall(ctx, &opts, i.pkgs...)
}
