#!/bin/bash
set -e

# Globals & sanity checks
IMG_PATH="$1"
BASE_PATH=${2%/}
SCRIPT_BASE_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

BOOT_IMG_MOUNTED_DEV=""
MAIN_IMG_MOUNTED_DEV=""
BOOT_IMG_MOUNT_POINT=""
MAIN_IMG_MOUNT_POINT=""
MAIN_PART_UUID=""

INITRD_SIZE_MB="128"

if [[ ! -d "${BASE_PATH}" ]]; then
  echo "Error: Base path directory '${BASE_PATH}' does not exist."
fi
if [[ ! -x /sbin/kpartx ]]; then
  echo "Error: kpartx is not installed."
  usage
  exit 1
fi

# Determine the kernel version
boot=`ls ${BASE_PATH}/boot/vmlinuz-*`
if [[ "$boot" != *"vmlinuz-"* ]]; then
  echo "Error: could not find linux image in ${BASE_PATH}/boot."
  exit 1
fi
boot="${boot##*/}"
KERN_IMG_FILENAME="$boot"
K_VERS=`echo $boot | sed -e "s/^vmlinuz-//" | cut -d"-" -f1`

# Functions

usage () {
  echo "USAGE: ./$0 <image-file-path> <base-build-path>"
}

mk_img_file () {
  echo "Creating image file '$1' with $2MB of space."
  dd if=/dev/zero of=$1 count=$2 bs=1M

  # Subtract INITRD_SIZE_MB for boot/initramfs partition.
  MAIN_SIZE="$(($2-${INITRD_SIZE_MB}))"
  PART_KIND="ext4"

  echo "Creating bootable ${PART_KIND} partition, using an msdos partition table."
  echo "Creating main partition ${MAIN_SIZE}MB in size."
  parted --script $1 mklabel msdos                                          \
                     mkpart p ext4 1 ${INITRD_SIZE_MB}                      \
                     mkpart p ${PART_KIND} ${INITRD_SIZE_MB} ${MAIN_SIZE}   \
                     set 1 boot on
}

mount_img () {
  losetup /dev/loop0 ${IMG_PATH}
  LOOP_INFO_RAW=`kpartx -av /dev/loop0`

  if [[ "${LOOP_INFO_RAW}" != *"add map "* ]]; then
    echo "Unexpected output from kpartx."
    echo "Output: ${LOOP_INFO_RAW}"
    kpartx -d ${IMG_PATH}
    exit 1
  fi

  echo "kpartx -av /dev/loop0: \"${LOOP_INFO_RAW}\""
  BOOT_IMG_MOUNTED_DEV=`echo ${LOOP_INFO_RAW} | head -n1 | awk -F"add map " '{$0=$2}1' | cut -d" " -f1`
  MAIN_IMG_MOUNTED_DEV=`echo ${LOOP_INFO_RAW} | tail -n1 | awk -F"add map " '{$0=$3}1' | cut -d" " -f1`
  echo "Boot partition mounted to ${BOOT_IMG_MOUNTED_DEV}"
  echo "Main partition mounted to ${MAIN_IMG_MOUNTED_DEV}"
}

unmount_img () {
  kpartx -d ${IMG_PATH}
  BOOT_IMG_MOUNTED_DEV=""
  MAIN_IMG_MOUNTED_DEV=""
  kpartx -v -d /dev/loop0
}

unmount_part () {
  umount /tmp/tmp_main_mnt
  rmdir /tmp/tmp_main_mnt
  MAIN_IMG_MOUNT_POINT=""
  if [[ "${BOOT_IMG_MOUNTED_DEV}" != "" ]]; then
    umount /tmp/tmp_boot_mnt
    rmdir /tmp/tmp_boot_mnt
    BOOT_IMG_MOUNT_POINT=""
  fi
}

mount_part () {
  mkdir /tmp/tmp_main_mnt || true
  mount /dev/mapper/${MAIN_IMG_MOUNTED_DEV} /tmp/tmp_main_mnt
  MAIN_IMG_MOUNT_POINT="/tmp/tmp_main_mnt"
  mkdir /tmp/tmp_boot_mnt || true
  mount /dev/mapper/${BOOT_IMG_MOUNTED_DEV} /tmp/tmp_boot_mnt
  BOOT_IMG_MOUNT_POINT="/tmp/tmp_boot_mnt"
}

on_error () {
  if [[ "${MAIN_IMG_MOUNT_POINT}" != "" ]]; then
    unmount_part
  fi
  if [[ "${MAIN_IMG_MOUNTED_DEV}" != "" ]]; then
    unmount_img
  fi
}

copy_files () {
  mkdir -p ${BOOT_IMG_MOUNT_POINT}/boot/grub

  cp -av "${BASE_PATH}/boot" "${BOOT_IMG_MOUNT_POINT}"
  cp -v "${SCRIPT_BASE_DIR}/grub_qemu.cfg" ${BOOT_IMG_MOUNT_POINT}/boot/grub/grub.cfg

  sed -i "s/KERN_IMG_FILENAME/${KERN_IMG_FILENAME}/" ${BOOT_IMG_MOUNT_POINT}/boot/grub/grub.cfg
  sed -i "s/K_VERS/${K_VERS}/" ${BOOT_IMG_MOUNT_POINT}/boot/grub/grub.cfg

  cp -av ${BASE_PATH}/bin ${MAIN_IMG_MOUNT_POINT}/bin
  cp -av ${BASE_PATH}/dev ${MAIN_IMG_MOUNT_POINT}/dev
  cp -av ${BASE_PATH}/etc ${MAIN_IMG_MOUNT_POINT}/etc
  cp -av ${BASE_PATH}/home ${MAIN_IMG_MOUNT_POINT}/home
  cp -av ${BASE_PATH}/lib ${MAIN_IMG_MOUNT_POINT}/lib
  cp -av ${BASE_PATH}/lib32 ${MAIN_IMG_MOUNT_POINT}/lib32
  cp -av ${BASE_PATH}/lib64 ${MAIN_IMG_MOUNT_POINT}/lib64
  cp -av ${BASE_PATH}/libx32 ${MAIN_IMG_MOUNT_POINT}/libx32
  cp -av ${BASE_PATH}/media ${MAIN_IMG_MOUNT_POINT}/media
  cp -av ${BASE_PATH}/mnt ${MAIN_IMG_MOUNT_POINT}/mnt
  cp -av ${BASE_PATH}/opt ${MAIN_IMG_MOUNT_POINT}/opt
  cp -av ${BASE_PATH}/proc ${MAIN_IMG_MOUNT_POINT}/proc
  cp -av ${BASE_PATH}/root ${MAIN_IMG_MOUNT_POINT}/root
  cp -av ${BASE_PATH}/run ${MAIN_IMG_MOUNT_POINT}/run
  cp -av ${BASE_PATH}/sbin ${MAIN_IMG_MOUNT_POINT}/sbin
  cp -av ${BASE_PATH}/srv ${MAIN_IMG_MOUNT_POINT}/srv
  cp -av ${BASE_PATH}/sys ${MAIN_IMG_MOUNT_POINT}/sys
  cp -av ${BASE_PATH}/tmp ${MAIN_IMG_MOUNT_POINT}/tmp
  cp -av ${BASE_PATH}/usr ${MAIN_IMG_MOUNT_POINT}/usr
  cp -av ${BASE_PATH}/var ${MAIN_IMG_MOUNT_POINT}/var
  cp -av ${BASE_PATH}/deb-pkgs ${MAIN_IMG_MOUNT_POINT}/deb-pkgs

  MAIN_PART_UUID=`lsblk -nr -o UUID /dev/mapper/${MAIN_IMG_MOUNTED_DEV}`
  BOOT_PART_UUID=`lsblk -nr -o UUID /dev/mapper/${BOOT_IMG_MOUNTED_DEV}`
  echo "UUID of main partition: $MAIN_PART_UUID"
  echo "UUID of boot partition: $BOOT_PART_UUID"
  sed -i "s/FSTAB_DEV/UUID=${MAIN_PART_UUID}/g" ${MAIN_IMG_MOUNT_POINT}/etc/fstab
  sed -i "s/BOOT_DEV/UUID=${BOOT_PART_UUID}/g" ${MAIN_IMG_MOUNT_POINT}/etc/fstab
  sed -i "s/MAIN_PART_UUID/${MAIN_PART_UUID}/g" ${BOOT_IMG_MOUNT_POINT}/boot/grub/grub.cfg
  sed -i "s/BOOT_PART_UUID/${BOOT_PART_UUID}/g" ${BOOT_IMG_MOUNT_POINT}/boot/grub/grub.cfg
}

install_grub () {
  DEV=/dev/${BOOT_IMG_MOUNTED_DEV::-2}
  echo "(hd0) ${DEV}" > /tmp/device.map
  grub-install -vvv --no-floppy                                                \
              --grub-mkdevicemap=/tmp/device.map                               \
              --modules="biosdisk part_msdos configfile normal multiboot"      \
              --root-directory=/tmp/tmp_boot_mnt                               \
              /dev/loop0
}


# Main code
trap 'on_error $LINENO' ERR EXIT

mk_img_file ${IMG_PATH} 12288

mount_img

sleep 2
mkfs.ext4 -q "/dev/mapper/${BOOT_IMG_MOUNTED_DEV}"
mkfs.ext4 -q "/dev/mapper/${MAIN_IMG_MOUNTED_DEV}"

mount_part
sleep 2

copy_files

install_grub

unmount_part
unmount_img
