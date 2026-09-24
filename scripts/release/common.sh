# shellcheck shell=bash
# shellcheck disable=SC2034 # biến dùng ở các script source file này
# Hàm dùng chung cho các script phát hành phần mềm desktop.
# Dùng: source "$(dirname "$0")/common.sh"

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DESKTOP_DIR="$REPO_ROOT/desktop"
BIN_DIR="$DESKTOP_DIR/build/bin"
# Thư mục chứa file phát hành (nằm trong build/bin nên đã bị gitignore).
OUT_DIR="${SANO_RELEASE_DIR:-$BIN_DIR/release}"
APP_NAME="Sano"

log() { printf '==> %s\n' "$*"; }
die() { printf 'LỖI: %s\n' "$*" >&2; exit 1; }

# Phiên bản hợp lệ: 1.2.3 hoặc 1.2.3-beta.1 hoặc 0.0.0-dev.abc1234.
check_version() {
  [[ "${1:-}" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ ]] \
    || die "phiên bản không hợp lệ: '${1:-}' (cần dạng 1.2.3 hoặc 1.2.3-beta.1)"
}

# Phần số X.Y.Z (Info.plist, thông tin file .exe chỉ nhận số).
numeric_version() {
  if [[ "$1" =~ ^([0-9]+\.[0-9]+\.[0-9]+) ]]; then
    printf '%s\n' "${BASH_REMATCH[1]}"
  else
    printf '0.0.0\n'
  fi
}

# Tên file phát hành: Sano-<version>-<os>-<arch>.<ext>
release_name() { printf '%s-%s-%s-%s.%s\n' "$APP_NAME" "$1" "$2" "$3" "$4"; }

sha256_of() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
  else
    shasum -a 256 "$1" | awk '{print $1}'
  fi
}

# Tải file về và kiểm SHA256 ghim sẵn; sai là dừng, xoá file.
fetch_verified() {
  local url="$1" want="$2" dest="$3" got
  log "Tải $url"
  curl -fsSL --retry 3 --proto '=https' -o "$dest" "$url"
  got="$(sha256_of "$dest")"
  if [[ "$got" != "$want" ]]; then
    rm -f "$dest"
    die "SHA256 không khớp cho $url: cần $want, nhận $got"
  fi
}
