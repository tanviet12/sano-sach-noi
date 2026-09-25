#!/usr/bin/env bash
# Tạo cặp khoá ed25519 ký bản cập nhật (làm MỘT lần, người giữ repo tự chạy).
#
#   scripts/release/update-key.sh [nơi lưu khoá bí mật]
#   # mặc định: ~/.sano/sano-update-signing-key.pem (quyền 600, ngoài repo)
#
# - Khoá công khai: ghi vào desktop/updatekey.go (commit dòng này).
# - Khoá bí mật: KHÔNG bao giờ commit. Đưa vào secret của repo rồi cất một bản
#   ở nơi an toàn (trình quản lý mật khẩu). Mất khoá = các bản đã cài không
#   nhận bản cập nhật ký bằng khoá mới → người dùng phải tải cài đè một lần.
#
# Đặt secret (GitHub CLI, không in khoá ra màn hình):
#   gh secret set SANO_UPDATE_SIGNING_KEY --repo <owner/repo> < <nơi lưu khoá>
# shellcheck source-path=SCRIPTDIR source=common.sh
source "$(dirname "$0")/common.sh"

KEY="${1:-$HOME/.sano/sano-update-signing-key.pem}"
[[ -e "$KEY" ]] && die "$KEY đã có — không ghi đè khoá cũ (xoá tay nếu thật sự muốn tạo lại)"
case "$(cd "$(dirname "$KEY")" 2>/dev/null && pwd)/" in
  "$REPO_ROOT"/*) die "không để khoá bí mật trong thư mục repo" ;;
esac
command -v openssl >/dev/null || die "thiếu openssl"

mkdir -p "$(dirname "$KEY")"
(umask 077 && openssl genpkey -algorithm ed25519 -out "$KEY")
PUB="$(openssl pkey -in "$KEY" -pubout -outform DER | tail -c 32 | openssl base64 -A)"
[[ ${#PUB} -eq 44 ]] || die "khoá công khai không đúng 32 byte"

GO="$DESKTOP_DIR/updatekey.go"
cur="$(sed -n 's/^var updatePublicKeys = "\(.*\)"$/\1/p' "$GO")"
new="$PUB"
[[ -n "$cur" ]] && new="$PUB,$cur" # đổi khoá: giữ khoá cũ vài bản rồi bỏ
sed -i.bak "s|^var updatePublicKeys = \".*\"$|var updatePublicKeys = \"$new\"|" "$GO" && rm -f "$GO.bak"

log "Khoá bí mật: $KEY (quyền 600)"
log "Khoá công khai: $PUB → đã ghi vào desktop/updatekey.go"
cat <<MSG

Bước tiếp:
  1. gh secret set SANO_UPDATE_SIGNING_KEY --repo tanviet12/sano-sach-noi < "$KEY"
  2. Cất một bản $KEY ở nơi an toàn (trình quản lý mật khẩu).
  3. Commit desktop/updatekey.go.
MSG
