#!/usr/bin/env bash
# Đóng gói Sano.app thành .dmg (kéo vào Applications).
#
#   scripts/release/package-macos.sh <version> [arch]   # arch mặc định: universal
#
# KHÔNG ký bằng chứng chỉ Apple, KHÔNG notarize (chưa đăng ký Apple Developer).
# Chỉ ký ad-hoc ("-") để máy Apple Silicon chạy được mã arm64 — không phải chữ
# ký nhận dạng, lần đầu mở vẫn cần "Vẫn mở" trong Cài đặt → Quyền riêng tư & Bảo mật.
# shellcheck source-path=SCRIPTDIR source=common.sh
source "$(dirname "$0")/common.sh"

VERSION="${1:-}"
ARCH="${2:-universal}"
check_version "$VERSION"

APP="$BIN_DIR/$APP_NAME.app"
[[ -d "$APP" ]] || die "không thấy $APP — chạy scripts/release/build.sh trước"

log "Ký ad-hoc $APP (không phải chữ ký Apple)"
codesign --force --deep --sign - "$APP"
codesign --verify --deep --strict "$APP"

mkdir -p "$OUT_DIR"
DMG="$OUT_DIR/$(release_name "$VERSION" macos "$ARCH" dmg)"
STAGE="$(mktemp -d)"
trap 'rm -rf "$STAGE"' EXIT

cp -R "$APP" "$STAGE/"
ln -s /Applications "$STAGE/Applications"

rm -f "$DMG"
log "hdiutil create $DMG"
hdiutil create -volname "$APP_NAME" -srcfolder "$STAGE" -fs HFS+ -format UDZO -ov "$DMG" >/dev/null
hdiutil verify -quiet "$DMG"
log "Xong: $DMG ($(du -h "$DMG" | awk '{print $1}'))"
