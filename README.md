# builder

 - Write the `/etc/os-release` & `/etc/issue` files.
 - Update the VERSION_MARKER in `/usr/sbin/twlinst-start`.

## Building TwitchyLinux

```shell
go get github.com/twitchylinux/builder
go build -o twl-builder github.com/twitchylinux/builder
```

### Build the root filesystem

```shell
sudo ./twl-builder /tmp/twitchylinux-fs
```

### Pack an image

```shell
sudo ./packscripts/pack_qemu.sh my-image.img /tmp/twitchylinux-fs
```


### Test in QEMU

```shell
sudo qemu-system-x86_64 -enable-kvm -cpu host -smp 2 -m 4G -drive format=raw,file=my-image.img
```
