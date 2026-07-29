#!/usr/bin/env bash
# Install bingo from the latest (or pinned) GitHub Release.
#
# No Go toolchain or git clone required — downloads the published binary
# for macOS Apple Silicon or Linux amd64.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/hotcuts/buzzword-bingo/main/scripts/install.sh | bash
#   # or from a checkout:
#   ./scripts/install.sh
#
# Env overrides:
#   REPO         default: hotcuts/buzzword-bingo
#   INSTALL_DIR  default: ~/.local/bin
#   VERSION      latest | vX.Y.Z  (default: latest)
#   ASSET        default: auto-detected (bingo_darwin_arm64 | bingo_linux_amd64)
#   NO_COLOR     set to disable ANSI colors
#
# Windows: use scripts/install.ps1 instead.

set -euo pipefail

REPO="${REPO:-hotcuts/buzzword-bingo}"
INSTALL_DIR="${INSTALL_DIR:-${HOME}/.local/bin}"
VERSION="${VERSION:-latest}"
ASSET="${ASSET:-}"
BINARY_NAME="bingo"
PLATFORM_LABEL=""

# Used by EXIT trap; must not be local to main (locals vanish before trap runs).
TMP_DOWNLOAD=""

# —— presentation ——
setup_colors() {
  if [[ -n "${NO_COLOR:-}" ]] || [[ ! -t 1 ]]; then
    C_TEAL=""
    C_MUTED=""
    C_OK=""
    C_ERR=""
    C_BOLD=""
    C_DIM=""
    C_RESET=""
    return
  fi
  C_TEAL=$'\033[38;2;45;212;191m'   # soft teal
  C_MUTED=$'\033[38;2;148;163;184m'  # slate
  C_OK=$'\033[38;2;52;211;153m'      # green
  C_ERR=$'\033[38;2;251;113;133m'    # rose
  C_BOLD=$'\033[1m'
  C_DIM=$'\033[2m'
  C_RESET=$'\033[0m'
}

banner() {
  printf '%s\n' "${C_TEAL}${C_BOLD}"
  cat <<'EOF'
██████╗ ██╗   ██╗███████╗███████╗██╗    ██╗ ██████╗ ██████╗ ██████╗
██╔══██╗██║   ██║╚══███╔╝╚══███╔╝██║    ██║██╔═══██╗██╔══██╗██╔══██╗
██████╔╝██║   ██║  ███╔╝   ███╔╝ ██║ █╗ ██║██║   ██║██████╔╝██║  ██║
██╔══██╗██║   ██║ ███╔╝   ███╔╝  ██║███╗██║██║   ██║██╔══██╗██║  ██║
██████╔╝╚██████╔╝███████╗███████╗╚███╔███╔╝╚██████╔╝██║  ██║██████╔╝
╚═════╝  ╚═════╝ ╚══════╝╚══════╝ ╚══╝╚══╝  ╚═════╝ ╚═╝  ╚═╝╚═════╝

██████╗ ██╗███╗   ██╗ ██████╗  ██████╗
██╔══██╗██║████╗  ██║██╔════╝ ██╔═══██╗
██████╔╝██║██╔██╗ ██║██║  ███╗██║   ██║
██╔══██╗██║██║╚██╗██║██║   ██║██║   ██║
██████╔╝██║██║ ╚████║╚██████╔╝╚██████╔╝
╚═════╝ ╚═╝╚═╝  ╚═══╝ ╚═════╝  ╚═════╝
EOF
  printf '%s' "${C_RESET}"
  printf '  %s%ssoft install · %s%s\n\n' "${C_MUTED}" "${C_DIM}" "${PLATFORM_LABEL:-detecting…}" "${C_RESET}"
}

step() {
  local label="$1"
  shift
  printf '  %s·%s %-8s %s%s%s\n' "${C_MUTED}" "${C_RESET}" "$label" "${C_MUTED}" "$*" "${C_RESET}"
}

ok() {
  local label="$1"
  shift
  printf '  %s✓%s %-8s %s\n' "${C_OK}" "${C_RESET}" "$label" "$*"
}

err() {
  printf '  %s✗%s %s\n' "${C_ERR}" "${C_RESET}" "$*" >&2
  exit 1
}

finale() {
  local version_line="$1"
  local playable="$2"
  local line1="Ready · ${version_line}"
  local line2
  if [[ "$playable" == "yes" ]]; then
    line2="Run     bingo play"
  else
    line2="Restart your shell, then bingo play"
  fi

  # Interior width: longest line + left/right padding (2 spaces each side).
  local pad=2
  local w=${#line1}
  (( ${#line2} > w )) && w=${#line2}
  w=$((w + pad * 2))

  local rule
  rule="$(printf '%*s' "$w" '' | tr ' ' '─')"
  local pad_l pad_r
  pad_l="$(printf '%*s' "$pad" '')"
  pad_r() {
    local content="$1"
    printf '%*s' $((w - pad - ${#content})) ''
  }

  printf '\n'
  printf '  %s╭%s╮%s\n' "${C_TEAL}" "$rule" "${C_RESET}"
  printf '  %s│%s%s%sReady%s · %s%s%s│%s\n' \
    "${C_TEAL}" "${C_RESET}" "$pad_l" "${C_BOLD}" "${C_RESET}" "$version_line" "$(pad_r "$line1")" "${C_TEAL}" "${C_RESET}"
  if [[ "$playable" == "yes" ]]; then
    printf '  %s│%s%sRun     %sbingo play%s%s%s│%s\n' \
      "${C_TEAL}" "${C_RESET}" "$pad_l" "${C_TEAL}${C_BOLD}" "${C_RESET}" "$(pad_r "$line2")" "${C_TEAL}" "${C_RESET}"
  else
    printf '  %s│%s%sRestart your shell, then %sbingo play%s%s%s│%s\n' \
      "${C_TEAL}" "${C_RESET}" "$pad_l" "${C_TEAL}${C_BOLD}" "${C_RESET}" "$(pad_r "$line2")" "${C_TEAL}" "${C_RESET}"
  fi
  printf '  %s╰%s╯%s\n' "${C_TEAL}" "$rule" "${C_RESET}"
}

cleanup() {
  rm -f "${TMP_DOWNLOAD:-}"
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || err "missing required command: $1"
}

# Sets PLATFORM_LABEL and ASSET (unless ASSET was already provided).
detect_platform() {
  local os arch detected_asset
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  arch="$(uname -m)"

  case "$os" in
    darwin)
      case "$arch" in
        arm64|aarch64)
          PLATFORM_LABEL="macOS arm64"
          detected_asset="bingo_darwin_arm64"
          ;;
        *)
          err "bingo on macOS only supports Apple Silicon (arm64), got: $arch"
          ;;
      esac
      ;;
    linux)
      case "$arch" in
        x86_64|amd64)
          PLATFORM_LABEL="Linux amd64"
          detected_asset="bingo_linux_amd64"
          ;;
        *)
          err "bingo on Linux only supports amd64, got: $arch"
          ;;
      esac
      ;;
    mingw*|msys*|cygwin*)
      err "use scripts/install.ps1 on Windows (got: $os)"
      ;;
    *)
      err "unsupported OS: $os (supported: macOS arm64, Linux amd64; Windows: install.ps1)"
      ;;
  esac

  if [[ -z "$ASSET" ]]; then
    ASSET="$detected_asset"
  fi
}

validate_binary() {
  local path="$1"
  local info
  info="$(file "$path")"

  case "$ASSET" in
    *darwin*)
      echo "$info" | grep -qi 'Mach-O' || \
        err "download is not a macOS binary — release asset ${ASSET} may be missing for ${VERSION}"
      ;;
    *linux*)
      echo "$info" | grep -qi 'ELF' || \
        err "download is not a Linux binary — release asset ${ASSET} may be missing for ${VERSION}"
      ;;
    *)
      # Custom ASSET override: accept Mach-O or ELF.
      echo "$info" | grep -qiE 'Mach-O|ELF' || \
        err "download is not a recognized Unix binary — release asset ${ASSET} may be missing for ${VERSION}"
      ;;
  esac
}

download_url() {
  local base="https://github.com/${REPO}/releases"
  if [[ "$VERSION" == "latest" ]]; then
    printf '%s/latest/download/%s' "$base" "$ASSET"
  else
    local tag="$VERSION"
    [[ "$tag" == v* ]] || tag="v${tag}"
    printf '%s/download/%s/%s' "$base" "$tag" "$ASSET"
  fi
}

ensure_dir_on_path() {
  local dir="$1"
  case ":${PATH}:" in
    *:"${dir}":*) return 0 ;;
  esac

  local line="export PATH=\"${dir}:\$PATH\""
  local rc="${ZDOTDIR:-$HOME}/.zshrc"
  case "${SHELL:-}" in
    */bash) rc="$HOME/.bashrc" ;;
  esac

  if [[ -f "$rc" ]] && grep -Fqs "$dir" "$rc"; then
    ok path "already in ${rc}"
  else
    step path "adding ${dir} → ${rc}"
    mkdir -p "$(dirname "$rc")"
    {
      printf '\n# bingo\n'
      printf '%s\n' "$line"
    } >>"$rc"
    ok path "added to ${rc}"
  fi

  export PATH="${dir}:${PATH}"
}

main() {
  setup_colors

  need_cmd uname
  need_cmd curl
  need_cmd mktemp
  need_cmd install
  need_cmd chmod
  need_cmd file
  need_cmd grep

  detect_platform
  banner
  ok detect "${PLATFORM_LABEL} → ${ASSET}"

  local url dest ver_msg playable
  url="$(download_url)"
  dest="${INSTALL_DIR}/${BINARY_NAME}"
  TMP_DOWNLOAD="$(mktemp)"
  trap cleanup EXIT

  step fetch "downloading release…"
  if ! curl -fsSL "$url" -o "$TMP_DOWNLOAD"; then
    err "download failed — is there a release with asset ${ASSET} on ${REPO}?"
  fi

  [[ -s "$TMP_DOWNLOAD" ]] || err "downloaded file is empty"
  validate_binary "$TMP_DOWNLOAD"
  ok fetch "${ASSET}"

  step install "writing ${dest}"
  chmod +x "$TMP_DOWNLOAD"
  mkdir -p "$INSTALL_DIR"
  install -m 0755 "$TMP_DOWNLOAD" "$dest"
  cleanup
  trap - EXIT
  TMP_DOWNLOAD=""
  ok install "${dest}"

  case ":${PATH}:" in
    *:"${INSTALL_DIR}":*)
      ok path "${INSTALL_DIR}"
      ;;
    *)
      ensure_dir_on_path "$INSTALL_DIR"
      ;;
  esac

  playable="no"
  ver_msg="${BINARY_NAME}"
  if command -v "$BINARY_NAME" >/dev/null 2>&1; then
    playable="yes"
    ver_msg="$($BINARY_NAME version 2>/dev/null || echo "$BINARY_NAME")"
  fi

  finale "$ver_msg" "$playable"
}

main "$@"
