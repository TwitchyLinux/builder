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

if [[ "${2}" == '--upgrade' ]]; then
  chroot $1 bash -c "cd \"/linux-${K_VERS}\" && make syncconfig"
else
  chroot $1 bash -c "cd \"/linux-${K_VERS}\" && make menuconfig"
fi
