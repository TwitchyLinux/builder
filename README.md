# builder

The builder builds a twitchylinux system. A modest amount of the image is configurable via
the files located in resources/.

## Building TwitchyLinux

Make sure you have [Go 1.13.4](https://golang.org/dl/) or greater installed.

```shell
go get github.com/twitchylinux/builder
go build -o twl-builder github.com/twitchylinux/builder
# The installer will now exist at ./twl-builder
```

### Build the root filesystem with kernel

```shell
sudo ./twl-builder /tmp/twitchylinux-fs
# Will create a twitchylinux system in /tmp/twitchylinux-fs
```

If you run this outside of the root of the `builder/` tree, you will need
to pass the path to the `resources/` directory, like this:

```shell
sudo ./twl-builder --resources-dir ~/builder/resources /tmp/twitchylinux-fs
```

### Write a LiveUSB

```shell
sudo packscripts/write_usb.sh /dev/sdd /tmp/twitchylinux-fs
# Assumes your USB is /dev/sdd
```

### Pack an image

```shell
sudo ./packscripts/pack_qemu.sh my-image.img /tmp/twitchylinux-fs
```


### Test in QEMU

**Execute image like a live CD**

```shell
xhost si:localuser:root
sudo qemu-system-x86_64 -soundhw hda -device virtio-rng-pci -vga virtio -enable-kvm -cpu host -smp 4 -m 4G -drive format=raw,file=my-image.img
```

**Create virtual drive & install**

```shell
qemu-img create -f qcow2 /tmp/qemu_hdd.img 25G # Make a hdd.
# Run once to install TwitchyLinux
sudo qemu-system-x86_64 -enable-kvm -serial stdio -vga qxl -cpu host -smp 4 -m 4G -drive format=raw,file=my-image.img -drive id=disk,file=/tmp/qemu_hdd.img,if=none -device ahci,id=ahci -device ide-drive,drive=disk,bus=ahci.0
# Run every time after to use
sudo qemu-system-x86_64 -enable-kvm -serial stdio -vga qxl -cpu host -smp 4 -m 4G -drive id=disk,file=/tmp/qemu_hdd.img,if=none -device ahci,id=ahci -device ide-drive,drive=disk,bus=ahci.0
```

## Notes on debugging sway

1. Make a temp rundir: `sudo mkdir /scratch/twl/tmp/xdg`
2. Bind mount your current rundir: `sudo mount --bind /run/user/1000/ /scratch/twl/tmp/xdg`
3. Launch the chroot with: `sudo chroot /scratch/twl sudo -u twl WLR_BACKENDS=wayland XDG_RUNTIME_DIR=/tmp/xdg sway -Vd`
