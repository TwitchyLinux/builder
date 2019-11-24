#!/bin/bash
set -e

# Globals
USB_PATH="$1"
BASE_PATH=${2%/}
SCRIPT_BASE_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

BOOT_IMG_MOUNT_POINT=""
MAIN_IMG_MOUNT_POINT=""

INITRD_SIZE_MB="128"
SYSINFO_SIZE_MB="512"



# Sanity checks.
# Check drive is a sane size.
LSBLK=`lsblk -d $1 | tail -n1 | tr -s ' '`
DEV_SIZE=`echo $LSBLK | cut -d" " -f4 | sed -e 's/[A-Z]*//g'`
echo "$1 has size of ${DEV_SIZE}G."
if (( $(echo "$DEV_SIZE > 64" | bc -l) )); then
  echo "ERROR: TwitchyLinux is intended for USB sticks 16-64Gb in size. Are you sure you have the right device?"
  exit 1
fi

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


#Find out if the USB is unmounted, and unmount it if so.
set +e
MNTS=`cat /proc/mounts | grep ${USB_PATH}`
set -e
if [[ "$MNTS" == *"$USB_PATH"* ]];then
  while read -r line; do
      PART=`echo $line | cut -d" " -f1`
      echo "Unmounting: $PART"
      umount $PART
  done <<< "$MNTS"
fi

format_drive () {
  # Subtract INITRD_SIZE_MB for boot/initramfs partition.
  DEV_SIZE_MB_FLOAT=$(echo "scale=4; $DEV_SIZE * 1024" | bc)
  DEV_SIZE_MB=$(echo "($DEV_SIZE_MB_FLOAT+0.5)/1" | bc )
  MAIN_SIZE="$(($DEV_SIZE_MB-${INITRD_SIZE_MB}-${SYSINFO_SIZE_MB}-256))"
  echo "Main partition will be ${MAIN_SIZE}Mb."

  MAIN_SIZE_OFFSET="$((${INITRD_SIZE_MB}+${MAIN_SIZE}))"
  SYSINFO_OFFSET="$((${MAIN_SIZE_OFFSET}+${SYSINFO_SIZE_MB}))"

  echo "Creating partition table..."
  parted --script $USB_PATH mklabel msdos                      \
         mkpart p fat32 1 ${INITRD_SIZE_MB}                    \
         mkpart p ext4  ${INITRD_SIZE_MB} ${MAIN_SIZE_OFFSET}  \
         mkpart p ext4  ${MAIN_SIZE_OFFSET} ${SYSINFO_OFFSET}  \
         set 1 boot on
  partprobe $USB_PATH
  sleep 2

  echo "Creating fat32 filesystem on ${USB_PATH}1..."
  mkfs.msdos -F 32 "${USB_PATH}1"
  echo "Creating ext4 filesystem on ${USB_PATH}2..."
  mkfs.ext4 -qF "${USB_PATH}2"
  echo "Creating ext4 filesystem on ${USB_PATH}3..."
  mkfs.ext4 -qF "${USB_PATH}3"

  sleep 2
  BOOT_PART_UUID=`lsblk -nr -o UUID ${USB_PATH}1`
  MAIN_PART_UUID=`lsblk -nr -o UUID ${USB_PATH}2`
  SYST_PART_UUID=`lsblk -nr -o UUID ${USB_PATH}3`
}

unmount_parts () {
  if [[ "${BOOT_IMG_MOUNT_POINT}" != "" ]]; then
    echo "Unmounting $BOOT_IMG_MOUNT_POINT"
    umount $BOOT_IMG_MOUNT_POINT
    BOOT_IMG_MOUNT_POINT=""
  fi
  if [[ "${MAIN_IMG_MOUNT_POINT}" != "" ]]; then
    echo "Unmounting $MAIN_IMG_MOUNT_POINT"
    echo "This may take multiple minutes due to the system page cache."
    echo "Do NOT remove the drive till this script exits or you WILL cause"
    echo "data corruption!"
    umount $MAIN_IMG_MOUNT_POINT
    MAIN_IMG_MOUNT_POINT=""
  fi
}

mount_parts () {
  mkdir -p /tmp/tmp_main_mnt
  mount "${USB_PATH}2" /tmp/tmp_main_mnt
  MAIN_IMG_MOUNT_POINT="/tmp/tmp_main_mnt"
  mkdir -p /tmp/tmp_boot_mnt || true
  mount "${USB_PATH}1" /tmp/tmp_boot_mnt
  BOOT_IMG_MOUNT_POINT="/tmp/tmp_boot_mnt"
}




copy_files () {
  mkdir -p ${BOOT_IMG_MOUNT_POINT}/boot/grub

  cp -av "${BASE_PATH}/boot" "${BOOT_IMG_MOUNT_POINT}"
  cp -v "${SCRIPT_BASE_DIR}/grub_usb.cfg" ${BOOT_IMG_MOUNT_POINT}/boot/grub/grub.cfg

  sed -i "s/KERN_IMG_FILENAME/${KERN_IMG_FILENAME}/g" ${BOOT_IMG_MOUNT_POINT}/boot/grub/grub.cfg
  sed -i "s/K_VERS/${K_VERS}/g" ${BOOT_IMG_MOUNT_POINT}/boot/grub/grub.cfg
  sed -i "s/BOOT_PART_UUID/${BOOT_PART_UUID}/g" ${BOOT_IMG_MOUNT_POINT}/boot/grub/grub.cfg
  sed -i "s/MAIN_PART_UUID/${MAIN_PART_UUID}/g" ${BOOT_IMG_MOUNT_POINT}/boot/grub/grub.cfg
  sed -i "s/SYST_PART_UUID/${SYST_PART_UUID}/g" ${BOOT_IMG_MOUNT_POINT}/boot/grub/grub.cfg

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

  sed -i "s/FSTAB_DEV/UUID=${MAIN_PART_UUID}/g" ${MAIN_IMG_MOUNT_POINT}/etc/fstab
  sed -i "s/BOOT_DEV/UUID=${BOOT_PART_UUID}/g" ${MAIN_IMG_MOUNT_POINT}/etc/fstab

  echo "Finished copying files."
}


install_grub () {
  echo "Installing grub..."
  echo "(hd0) ${USB_PATH}" > /tmp/device.map
  grub-install -vvv --no-floppy                                                \
              --grub-mkdevicemap=/tmp/device.map                               \
              --modules="biosdisk part_msdos configfile normal multiboot"      \
              --root-directory=/tmp/tmp_boot_mnt                               \
              ${USB_PATH}
  echo "Finished installing grub."
}





# Main code
trap 'unmount_parts' ERR EXIT

format_drive

mount_parts
sleep 1

copy_files
install_grub

unmount_parts
