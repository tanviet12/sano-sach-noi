#!/usr/bin/env bash
# Build phần mềm desktop bằng Wails, gắn phiên bản vào app.
#
#   scripts/release/build.sh <version> [platform]
#
# platform mặc định theo máy đang chạy: darwin/universal, windows/amd64, linux/amd64.
# Windows: tạo thêm bộ cài NSIS cài theo người dùng (không đòi quyền admin).
# Linux: build với -tags webkit2_41 (cần libwebkit2gtk-4.1-dev).
# Không nén UPX (dễ bị antivirus báo nhầm).
#
# Tạm ghi productVersion vào desktop/wails.json (Info.plist, thông tin file
# .exe, bộ cài), build xong trả lại nguyên bản.
# shellcheck source-path=SCRIPTDIR source=common.sh
source "$(dirname "$0")/common.sh"

VERSION="${1:-}"
check_version "$VERSION"

PLATFORM="${2:-}"
if [[ -z "$PLATFORM" ]]; then
  case "$(uname -s)" in
    Darwin) PLATFORM="darwin/universal" ;;
    Linux) PLATFORM="linux/amd64" ;;
    MINGW* | MSYS* | CYGWIN*) PLATFORM="windows/amd64" ;;
    *) die "không nhận ra hệ điều hành: $(uname -s)" ;;
  esac
fi

WAILS="${WAILS:-$(command -v wails || true)}"
[[ -n "$WAILS" ]] || die "chưa có Wails CLI (go install github.com/wailsapp/wails/v2/cmd/wails@<phiên bản trong desktop/go.mod>)"

args=(build -clean -trimpath -m -platform "$PLATFORM" -ldflags "-X main.version=$VERSION")
case "$PLATFORM" in
  linux/*) args+=(-tags webkit2_41) ;;
  windows/*) args+=(-nsis -installscope user -webview2 download) ;;
esac

WAILS_JSON="$DESKTOP_DIR/wails.json"
BACKUP="$(mktemp)"
cp "$WAILS_JSON" "$BACKUP"
restore() { cp "$BACKUP" "$WAILS_JSON"; rm -f "$BACKUP"; }
trap restore EXIT

NUM="$(numeric_version "$VERSION")"
log "productVersion = $NUM (phiên bản app: $VERSION)"
WAILS_JSON="$WAILS_JSON" NUM="$NUM" node -e '
  const fs = require("fs");
  const p = process.env.WAILS_JSON;
  const j = JSON.parse(fs.readFileSync(p, "utf8"));
  j.info = j.info || {};
  j.info.productVersion = process.env.NUM;
  fs.writeFileSync(p, JSON.stringify(j, null, 2) + "\n");
'

log "wails ${args[*]}"
(cd "$DESKTOP_DIR" && "$WAILS" "${args[@]}")
log "Xong: $(ls "$BIN_DIR")"
