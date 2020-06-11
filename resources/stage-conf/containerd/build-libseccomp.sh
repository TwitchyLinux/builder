#!/bin/bash

set -eu -o pipefail

set -x

export SECCOMP_VERSION="2.4.1"
export SECCOMP_PATH="$(mktemp -d)"
curl -fsSL "https://github.com/seccomp/libseccomp/releases/download/v${SECCOMP_VERSION}/libseccomp-${SECCOMP_VERSION}.tar.gz" | tar -xzC "$SECCOMP_PATH" --strip-components=1
(
	cd "$SECCOMP_PATH"
	./configure --prefix=/usr/local
	make
	make install
	ldconfig
)

rm -rf "$SECCOMP_PATH"
