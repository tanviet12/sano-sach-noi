#!/usr/bin/env bash
# Đóng gói Sano.app thành .dmg (kéo vào Applications), ký Developer ID + notarize.
#
#   scripts/release/package-macos.sh <version> [arch]   # arch mặc định: universal
#
# Ký số Apple (Developer ID Application):
#   SANO_MAC_SIGN_IDENTITY  tên chứng chỉ; bỏ trống thì tự tìm "Developer ID Application"
#                           trong Keychain (SANO_MAC_KEYCHAIN nếu có).
#   Không có chứng chỉ → ký ad-hoc như trước (người dùng phải bấm "Vẫn mở"),
#   trừ khi SANO_REQUIRE_MAC_SIGN=1 (bản phát hành) thì dừng.
#
# Notarize (chỉ khi đã ký Developer ID), một trong hai cách:
#   SANO_NOTARY_PROFILE     hồ sơ đã lưu bằng "xcrun notarytool store-credentials" (máy anh: sano-notary)
#   SANO_NOTARY_KEY + SANO_NOTARY_KEY_ID + SANO_NOTARY_ISSUER
#                           nội dung file AuthKey_*.p8 + Key ID + Issuer ID (GitHub Actions)
#   Có chứng chỉ mà không có thông tin notarize → dừng nếu SANO_REQUIRE_MAC_SIGN=1.
# shellcheck source-path=SCRIPTDIR source=common.sh
source "$(dirname "$0")/common.sh"

VERSION="${1:-}"
ARCH="${2:-universal}"
check_version "$VERSION"

APP="$BIN_DIR/$APP_NAME.app"
[[ -d "$APP" ]] || die "không thấy $APP — chạy scripts/release/build.sh trước"

REQUIRE="${SANO_REQUIRE_MAC_SIGN:-0}"
KEYCHAIN_ARGS=()
[[ -n "${SANO_MAC_KEYCHAIN:-}" ]] && KEYCHAIN_ARGS=(--keychain "$SANO_MAC_KEYCHAIN")

IDENTITY="${SANO_MAC_SIGN_IDENTITY:-}"
if [[ -z "$IDENTITY" ]]; then
  IDENTITY="$(security find-identity -v -p codesigning ${SANO_MAC_KEYCHAIN:+"$SANO_MAC_KEYCHAIN"} 2>/dev/null \
    | sed -n 's/.*"\(Developer ID Application: .*\)"$/\1/p' | head -n1)"
fi

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

if [[ -n "$IDENTITY" ]]; then
  log "Ký Developer ID: $IDENTITY"
  sign() { codesign --force --options runtime --timestamp --sign "$IDENTITY" ${KEYCHAIN_ARGS[@]+"${KEYCHAIN_ARGS[@]}"} "$@"; }
  # Ký từ trong ra ngoài: file chạy phụ trước, cả gói .app sau cùng.
  sign "$APP/Contents/MacOS/sano-mcp"
  sign "$APP"
  codesign --verify --deep --strict --verbose=2 "$APP"
else
  [[ "$REQUIRE" == 1 ]] && die "không thấy chứng chỉ Developer ID Application (SANO_REQUIRE_MAC_SIGN=1)"
  log "Không có chứng chỉ Developer ID → ký ad-hoc (không phải chữ ký Apple)"
  codesign --force --deep --sign - "$APP"
  codesign --verify --deep --strict "$APP"
fi

mkdir -p "$OUT_DIR"
DMG="$OUT_DIR/$(release_name "$VERSION" macos "$ARCH" dmg)"
STAGE="$WORK/stage"
mkdir -p "$STAGE"

cp -R "$APP" "$STAGE/"
ln -s /Applications "$STAGE/Applications"

rm -f "$DMG"
log "hdiutil create $DMG"
# Máy Mac của GitHub Actions đôi khi báo "hdiutil: create failed - Resource
# busy" (ổ ảnh tạm còn bị giữ) → thử lại vài lần.
ok=false
for try in 1 2 3; do
  if hdiutil create -volname "$APP_NAME" -srcfolder "$STAGE" -fs HFS+ -format UDZO -ov "$DMG" >/dev/null; then
    ok=true
    break
  fi
  log "hdiutil create lỗi (lần $try), thử lại sau 10 giây"
  rm -f "$DMG"
  sleep 10
done
$ok || die "hdiutil create lỗi 3 lần"
hdiutil verify -quiet "$DMG"

if [[ -n "$IDENTITY" ]]; then
  codesign --force --timestamp --sign "$IDENTITY" ${KEYCHAIN_ARGS[@]+"${KEYCHAIN_ARGS[@]}"} "$DMG"

  NOTARY=()
  if [[ -n "${SANO_NOTARY_PROFILE:-}" ]]; then
    NOTARY=(--keychain-profile "$SANO_NOTARY_PROFILE")
  elif [[ -n "${SANO_NOTARY_KEY:-}" && -n "${SANO_NOTARY_KEY_ID:-}" && -n "${SANO_NOTARY_ISSUER:-}" ]]; then
    (umask 077 && printf '%s\n' "$SANO_NOTARY_KEY" >"$WORK/notary.p8")
    NOTARY=(--key "$WORK/notary.p8" --key-id "$SANO_NOTARY_KEY_ID" --issuer "$SANO_NOTARY_ISSUER")
  fi

  if [[ ${#NOTARY[@]} -gt 0 ]]; then
    # Notarize file .dmg (Apple kiểm luôn Sano.app bên trong), rồi gắn vé duyệt vào .dmg
    # để máy không có mạng vẫn mở được lần đầu.
    log "Gửi Apple notarize (thường 2–15 phút)"
    xcrun notarytool submit "$DMG" "${NOTARY[@]}" --wait --timeout 60m --output-format json >"$WORK/notary.json" || true
    status="$(sed -n 's/.*"status" *: *"\([^"]*\)".*/\1/p' "$WORK/notary.json" | head -n1)"
    id="$(sed -n 's/.*"id" *: *"\([^"]*\)".*/\1/p' "$WORK/notary.json" | head -n1)"
    if [[ "$status" != "Accepted" ]]; then
      cat "$WORK/notary.json" >&2
      [[ -n "$id" ]] && xcrun notarytool log "$id" "${NOTARY[@]}" >&2 || true
      die "notarize không đạt (trạng thái: ${status:-không rõ})"
    fi
    log "Notarize đạt (mã $id)"
    xcrun stapler staple "$DMG"
    xcrun stapler validate "$DMG"
    spctl --assess --type open --context context:primary-signature --verbose=2 "$DMG"
  else
    [[ "$REQUIRE" == 1 ]] && die "thiếu thông tin notarize (SANO_NOTARY_PROFILE hoặc SANO_NOTARY_KEY/_KEY_ID/_ISSUER)"
    log "Chưa notarize (không có thông tin notarytool) — người dùng vẫn bị Gatekeeper hỏi lần đầu"
  fi
fi

log "Xong: $DMG ($(du -h "$DMG" | awk '{print $1}'))"
