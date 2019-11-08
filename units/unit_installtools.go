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
			"arping", "arptables", "dnsutils",
		},
	}
	compressionToolsInstall = &InstallTools{
		name: "compression-tools",
		pkgs: []string{"lzip", "unzip", "zip", "tar", "rar", "unrar"},
	}
	cliToolsInstall = &InstallTools{
		name: "cli-tools",
		pkgs: []string{
			"htop", "screen", "nano", "vim", "vim-doc", "git", "git-doc",
			"figlet",
			"sharutils", "sharutils-doc",
			"pm-utils", "pciutils", "sysstat", "dpkg-dev",
			"gnupg", "jq", "tidy", "dstat", "info", "attr",
		},
	}
	usbInstall = &InstallTools{
		name: "usb",
		pkgs: []string{"usbutils", "libusb-1.0-0", "libusb-1.0-0-dev", "libusb-1.0-doc"},
	}
	cToolchainInstall = &InstallTools{
		name: "c-toolchain",
		pkgs: []string{"build-essential", "cmake", "sqlite3", "libsqlite3-0", "libsqlite3-dev", "sqlite3-doc"},
	}
)

type InstallTools struct {
	name string
	pkgs []string
}

func (i *InstallTools) Name() string {
	return i.name
}

func (i *InstallTools) Run(ctx context.Context, opts Opts) error {
	chroot, err := prepareChroot(opts.Dir)
	if err != nil {
		return err
	}
	defer chroot.Close()

	return chroot.AptInstall(ctx, &opts, i.pkgs...)
}
