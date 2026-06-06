#!/usr/bin/env bash
# GitHub Releases から statecast バイナリをダウンロードしてインストールする
set -euo pipefail

REPO="${STATECAST_REPO:-okonomipizza/statecast}"
VERSION="${VERSION:-}"
INSTALL_DIR="${INSTALL_DIR:-${HOME}/.local/bin}"

log() {
  printf '%s\n' "$*" >&2
}

die() {
  log "error: $*"
  exit 1
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "required command not found: $1"
}

detect_platform() {
  local os arch

  os="$(uname -s)"
  arch="$(uname -m)"

  case "${os}" in
    Darwin) os="darwin" ;;
    Linux) os="linux" ;;
    *) die "unsupported OS: ${os} (macOS and Linux only)" ;;
  esac

  case "${arch}" in
    x86_64 | amd64) arch="amd64" ;;
    arm64 | aarch64) arch="arm64" ;;
    *) die "unsupported architecture: ${arch}" ;;
  esac

  printf '%s %s' "${os}" "${arch}"
}

resolve_version() {
  if [[ -n "${VERSION}" ]]; then
    if [[ "${VERSION}" != v* ]]; then
      VERSION="v${VERSION}"
    fi
    printf '%s' "${VERSION}"
    return
  fi

  need_cmd curl
  need_cmd grep
  need_cmd sed

  curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name":' \
    | sed -E 's/.*"tag_name": "([^"]+)".*/\1/' \
    || die "failed to resolve latest release version"
}

verify_checksum() {
  local file="$1"
  local expected
  expected="$(grep "${file}" checksums.txt | awk '{print $1}')"

  if [[ -z "${expected}" ]]; then
    die "checksum entry not found for ${file}"
  fi

  local actual
  if command -v sha256sum >/dev/null 2>&1; then
    actual="$(sha256sum "${file}" | awk '{print $1}')"
  elif command -v shasum >/dev/null 2>&1; then
    actual="$(shasum -a 256 "${file}" | awk '{print $1}')"
  else
    log "warning: sha256sum/shasum not found, skipping verification"
    return 0
  fi

  if [[ "${actual}" != "${expected}" ]]; then
    die "checksum mismatch: expected ${expected}, got ${actual}"
  fi
}

main() {
  need_cmd curl
  need_cmd tar
  need_cmd mktemp

  read -r os arch <<<"$(detect_platform)"
  tag="$(resolve_version)"
  version="${tag#v}"
  asset="statecast_${version}_${os}_${arch}.tar.gz"
  url="https://github.com/${REPO}/releases/download/${tag}/${asset}"

  tmpdir="$(mktemp -d)"
  trap 'rm -rf "${tmpdir}"' EXIT

  log "installing statecast ${tag} for ${os}/${arch}"
  log "downloading ${url}"

  curl -fsSL "${url}" -o "${tmpdir}/${asset}"

  checksums_url="https://github.com/${REPO}/releases/download/${tag}/checksums.txt"
  if curl -fsSL "${checksums_url}" -o "${tmpdir}/checksums.txt" 2>/dev/null; then
    log "verifying checksum"
    (cd "${tmpdir}" && verify_checksum "${asset}")
  else
    log "warning: checksums.txt not found, skipping verification"
  fi

  tar -xzf "${tmpdir}/${asset}" -C "${tmpdir}"

  mkdir -p "${INSTALL_DIR}"
  install -m 755 "${tmpdir}/statecast" "${INSTALL_DIR}/statecast"

  log "installed to ${INSTALL_DIR}/statecast"
  if ! printf '%s\n' "${PATH}" | tr ':' '\n' | grep -qx "${INSTALL_DIR}"; then
    log "hint: add ${INSTALL_DIR} to PATH"
    log "  export PATH=\"${INSTALL_DIR}:\$PATH\""
  fi
}

main "$@"
