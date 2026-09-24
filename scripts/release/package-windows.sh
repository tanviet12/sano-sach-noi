#!/usr/bin/env bash
# Đổi tên bộ cài NSIS + nén bản chạy ngay (portable) cho Windows.
#
#   scripts/release/package-windows.sh <version> [arch]   # arch mặc định: amd64
#
# Ra: Sano-<version>-windows-<arch>-setup.exe (cài theo người dùng, không cần
# admin) và Sano-<version>-windows-<arch>-portable.zip (chỉ Sano.exe).
# Chưa ký số: lần đầu chạy SmartScreen báo → "Thông tin thêm" → "Vẫn chạy".
# shellcheck source-path=SCRIPTDIR source=common.sh
source "$(dirname "$0")/common.sh"

VERSION="${1:-}"
ARCH="${2:-amd64}"
check_version "$VERSION"

EXE="$BIN_DIR/$APP_NAME.exe"
INSTALLER="$BIN_DIR/$APP_NAME-$ARCH-installer.exe"
[[ -f "$EXE" ]] || die "không thấy $EXE — chạy scripts/release/build.sh trước"
[[ -f "$INSTALLER" ]] || die "không thấy $INSTALLER — thiếu makensis? (bộ cài NSIS)"

mkdir -p "$OUT_DIR"
SETUP="$OUT_DIR/$(release_name "$VERSION" windows "$ARCH" exe)"
SETUP="${SETUP%.exe}-setup.exe"
ZIP="$OUT_DIR/$(release_name "$VERSION" windows "$ARCH" zip)"
ZIP="${ZIP%.zip}-portable.zip"

cp "$INSTALLER" "$SETUP"
log "Bộ cài: $SETUP"

rm -f "$ZIP"
if command -v 7z >/dev/null 2>&1; then
  (cd "$BIN_DIR" && 7z a -tzip -bd "$ZIP" "$APP_NAME.exe" >/dev/null)
elif command -v zip >/dev/null 2>&1; then
  (cd "$BIN_DIR" && zip -q "$ZIP" "$APP_NAME.exe")
else
  powershell.exe -NoProfile -Command "Compress-Archive -Path '$(cygpath -w "$EXE")' -DestinationPath '$(cygpath -w "$ZIP")'"
fi
log "Bản chạy ngay: $ZIP"
