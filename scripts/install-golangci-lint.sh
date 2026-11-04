#!/bin/sh
# Install an official golangci-lint release binary after verifying its SHA256.
#
# Usage: scripts/install-golangci-lint.sh <version> <bindir>
#   e.g. scripts/install-golangci-lint.sh v2.15.0 "$(go env GOPATH)/bin"
#
# Why not upstream install.sh: since v2.12.0 the release checksums file also
# lists *.tar.gz.sbom.json, and install.sh looks up the expected hash with a
# substring grep that matches both lines, so verification always fails.
# This script matches the archive file name exactly.
set -eu

if [ "$#" -ne 2 ]; then
	echo "usage: $0 <version> <bindir>" >&2
	exit 2
fi

version=$1
bindir=$2
num=${version#v}

case "$(uname -s)" in
	Linux) os=linux ;;
	Darwin) os=darwin ;;
	*) echo "unsupported OS: $(uname -s)" >&2; exit 1 ;;
esac

case "$(uname -m)" in
	x86_64 | amd64) arch=amd64 ;;
	aarch64 | arm64) arch=arm64 ;;
	*) echo "unsupported architecture: $(uname -m)" >&2; exit 1 ;;
esac

name="golangci-lint-${num}-${os}-${arch}"
archive="${name}.tar.gz"
checksums="golangci-lint-${num}-checksums.txt"
base_url="https://github.com/golangci/golangci-lint/releases/download/${version}"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

curl -sSfL -o "${tmp}/${archive}" "${base_url}/${archive}"
curl -sSfL -o "${tmp}/${checksums}" "${base_url}/${checksums}"

want=$(awk -v f="$archive" '$2 == f { print $1 }' "${tmp}/${checksums}")
if [ -z "$want" ]; then
	echo "no checksum for ${archive} in ${checksums}" >&2
	exit 1
fi

if command -v sha256sum >/dev/null 2>&1; then
	got=$(sha256sum "${tmp}/${archive}" | awk '{ print $1 }')
else
	got=$(shasum -a 256 "${tmp}/${archive}" | awk '{ print $1 }')
fi

if [ "$want" != "$got" ]; then
	echo "checksum mismatch for ${archive}: want ${want}, got ${got}" >&2
	exit 1
fi

tar -xzf "${tmp}/${archive}" -C "$tmp" "${name}/golangci-lint"
mkdir -p "$bindir"
install -m 0755 "${tmp}/${name}/golangci-lint" "${bindir}/golangci-lint"
echo "installed golangci-lint ${version} to ${bindir}/golangci-lint"
