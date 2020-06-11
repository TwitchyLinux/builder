#!/bin/bash

if [ ! -f "/usr/local/lib/libseccomp.so.2.4.1" ]; then
  bash "$(dirname $0)/build-libseccomp.sh"
fi

set -eu -o pipefail
set -x

export BIN_PATH='/usr/local/containerd/bin'
export CONTAINERD_VERSION="1.4.0-beta.1"
export BUILD_PATH="$(mktemp -d)"
trap "rm -rf ${BUILD_PATH}" SIGINT SIGTERM EXIT

curl -fsSL "https://github.com/containerd/containerd/releases/download/v${CONTAINERD_VERSION}/containerd-${CONTAINERD_VERSION}-linux-amd64.tar.gz" | tar -xzC "$BUILD_PATH" --strip-components=1
(
  if [ -d "${BIN_PATH}" ]; then
    rm -rfv "$BIN_PATH"
  fi
  mkdir -pv "$BIN_PATH"
  cp "${BUILD_PATH}"/* /usr/local/containerd/bin
)

cat << 'EOF' > /etc/profile.d/containerd.sh
export PATH="/usr/local/containerd/bin:$PATH"
EOF
chmod +x /etc/profile.d/containerd.sh
. /etc/profile.d/containerd.sh


cat << 'EOF' > /lib/systemd/system/containerd.service
[Unit]
Description=containerd container runtime
Documentation=https://containerd.io
After=network.target local-fs.target

[Service]
ExecStartPre=-/sbin/modprobe overlay
ExecStart=/usr/local/containerd/bin/containerd

Type=notify
Delegate=yes
KillMode=process
Restart=always
# Having non-zero Limit*s causes performance problems due to accounting overhead
# in the kernel. We recommend using cgroups to do container-local accounting.
LimitNPROC=infinity
LimitCORE=infinity
LimitNOFILE=1048576
# Comment TasksMax if your systemd version does not supports it.
# Only systemd 226 and above support this version.
TasksMax=infinity

[Install]
WantedBy=multi-user.target
EOF

mkdir -pv /etc/containerd
containerd config default > /etc/containerd/config.toml
