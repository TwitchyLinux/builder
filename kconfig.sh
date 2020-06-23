#!/bin/bash
# ./kconfig.sh <path-to-almost-built-system>

SCRIPT_BASE_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

#Determine linux version from the sources.
LIN=`ls $1/linux-*`
LIN="${LIN##*/}"
K_VERS=`echo $LIN | sed -e "s/^linux-//" | cut -d" " -f1| sed -e "s/://"`

# Prepare source tree
chroot $1 bash -c "cd \"/linux-${K_VERS}\" && make mrproper"

cp -v "${SCRIPT_BASE_DIR}/resources/linux/.config" "${1}/linux-${K_VERS}/.config"



case "${2}" in
  --upgrade)
    chroot $1 bash -c "cd \"/linux-${K_VERS}\" && make syncconfig"
    ;;
  --enable)
    chroot $1 bash -c "cd \"/linux-${K_VERS}\" && scripts/config --enable ${3}"
    chroot $1 bash -c "cd \"/linux-${K_VERS}\" && make olddefconfig"
    ;;
  --module)
    chroot $1 bash -c "cd \"/linux-${K_VERS}\" && scripts/config --module ${3}"
    chroot $1 bash -c "cd \"/linux-${K_VERS}\" && make olddefconfig"
    ;;
  --disable)
    chroot $1 bash -c "cd \"/linux-${K_VERS}\" && scripts/config --disable ${3}"
    chroot $1 bash -c "cd \"/linux-${K_VERS}\" && make olddefconfig"
    ;;
  *)
    chroot $1 bash -c "cd \"/linux-${K_VERS}\" && make menuconfig"
    ;;
esac
