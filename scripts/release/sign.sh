#!/usr/bin/env bash
# Ký SHA256SUMS bằng khoá ed25519 riêng của dự án → SHA256SUMS.sig (base64 của
# chữ ký 64 byte, một dòng). App kiểm chữ ký này bằng khoá công khai nhúng sẵn
# (desktop/updatekey.go) trước khi tự cập nhật.
#
#   SANO_UPDATE_SIGNING_KEY="$(cat khoa.pem)" scripts/release/sign.sh [thư mục]
#
# Khoá bí mật (PEM PKCS#8) chỉ đọc từ biến môi trường, ghi tạm vào file quyền
# 600 rồi xoá. Ký xong kiểm lại bằng khoá công khai trong desktop/updatekey.go:
# khoá trong secret không khớp khoá nhúng trong app thì dừng (app sẽ từ chối).
# shellcheck source-path=SCRIPTDIR source=common.sh
source "$(dirname "$0")/common.sh"

DIR="${1:-$OUT_DIR}"
[[ -f "$DIR/SHA256SUMS" ]] || die "không thấy $DIR/SHA256SUMS — chạy checksums.sh trước"
[[ -n "${SANO_UPDATE_SIGNING_KEY:-}" ]] || die "thiếu SANO_UPDATE_SIGNING_KEY (secret của repo, xem scripts/release/update-key.sh)"
command -v openssl >/dev/null || die "thiếu openssl"

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
(umask 077 && printf '%s\n' "$SANO_UPDATE_SIGNING_KEY" >"$WORK/key.pem")

openssl pkeyutl -sign -rawin -inkey "$WORK/key.pem" -in "$DIR/SHA256SUMS" -out "$WORK/sig.bin" \
  || die "ký thất bại (khoá không phải ed25519 PEM?)"
[[ "$(wc -c <"$WORK/sig.bin" | tr -d ' ')" == 64 ]] || die "chữ ký không đủ 64 byte"
rm -f "$DIR/SHA256SUMS.sig" # chỉ ghi chữ ký mới sau khi kiểm lại đạt

# Kiểm lại bằng các khoá công khai nhúng trong app.
keys="$(sed -n 's/^var updatePublicKeys = "\(.*\)"$/\1/p' "$DESKTOP_DIR/updatekey.go")"
[[ -n "$keys" ]] || die "desktop/updatekey.go chưa có khoá công khai — app sẽ không tự cập nhật được"
ok=false
IFS=',' read -ra list <<<"$keys"
for k in "${list[@]}"; do
  k="${k// /}"
  [[ -n "$k" ]] || continue
  # SubjectPublicKeyInfo của ed25519 = tiền tố DER cố định + 32 byte khoá.
  { printf '\x30\x2a\x30\x05\x06\x03\x2b\x65\x70\x03\x21\x00'; printf '%s' "$k" | openssl base64 -d -A; } >"$WORK/pub.der"
  if openssl pkeyutl -verify -rawin -pubin -keyform DER -inkey "$WORK/pub.der" \
    -in "$DIR/SHA256SUMS" -sigfile "$WORK/sig.bin" >/dev/null 2>&1; then
    ok=true
    break
  fi
done
$ok || die "chữ ký không khớp khoá công khai nào trong desktop/updatekey.go — secret và khoá nhúng trong app lệch nhau"
{ openssl base64 -A -in "$WORK/sig.bin"; printf '\n'; } >"$DIR/SHA256SUMS.sig"
log "Đã ký: $DIR/SHA256SUMS.sig (khớp khoá trong desktop/updatekey.go)"
