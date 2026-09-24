#!/usr/bin/env bash
# Đóng gói bản Linux thành .AppImage.
#
#   scripts/release/package-linux.sh <version> [arch]   # arch mặc định: amd64
#
# AppImage "gọn": chỉ chứa Sano + biểu tượng, dùng GTK3 + WebKitGTK 4.1 có sẵn
# trên máy (Ubuntu 22.04+, Debian 12+, Fedora 36+...). Không nhét WebKitGTK vào
# gói vì tiến trình con của WebKit bị gắn đường dẫn cố định, nhét vào dễ hỏng.
# Runtime type2 tĩnh: không cần libfuse2.
#
# appimagetool + runtime tải đúng bản ghim, kiểm SHA256 (lấy từ GitHub Release,
# đối chiếu trường digest của GitHub).
# shellcheck source-path=SCRIPTDIR source=common.sh
source "$(dirname "$0")/common.sh"

APPIMAGETOOL_VERSION="1.9.1"
APPIMAGETOOL_SHA256="ed4ce84f0d9caff66f50bcca6ff6f35aae54ce8135408b3fa33abfc3cb384eb0"
RUNTIME_VERSION="20251108"
RUNTIME_SHA256="2fca8b443c92510f1483a883f60061ad09b46b978b2631c807cd873a47ec260d"

VERSION="${1:-}"
ARCH="${2:-amd64}"
check_version "$VERSION"
[[ "$ARCH" == "amd64" ]] || die "hiện chỉ đóng gói amd64 (x86_64)"

BIN="$BIN_DIR/$APP_NAME"
[[ -f "$BIN" ]] || die "không thấy $BIN — chạy scripts/release/build.sh trước"

TOOLS="${SANO_TOOLS_DIR:-$DESKTOP_DIR/build/bin/.tools}"
mkdir -p "$TOOLS"
TOOL="$TOOLS/appimagetool-$APPIMAGETOOL_VERSION-x86_64.AppImage"
RUNTIME="$TOOLS/runtime-$RUNTIME_VERSION-x86_64"
[[ -f "$TOOL" && "$(sha256_of "$TOOL")" == "$APPIMAGETOOL_SHA256" ]] || fetch_verified \
  "https://github.com/AppImage/appimagetool/releases/download/$APPIMAGETOOL_VERSION/appimagetool-x86_64.AppImage" \
  "$APPIMAGETOOL_SHA256" "$TOOL"
[[ -f "$RUNTIME" && "$(sha256_of "$RUNTIME")" == "$RUNTIME_SHA256" ]] || fetch_verified \
  "https://github.com/AppImage/type2-runtime/releases/download/$RUNTIME_VERSION/runtime-x86_64" \
  "$RUNTIME_SHA256" "$RUNTIME"
# Bản chạy: xoá "AI\x02" ở byte 8–10 của header ELF (dấu AppImage) trên bản
# sao đã kiểm SHA256 — không ảnh hưởng mã chạy, nhưng giúp chạy được qua bộ
# giả lập (Docker amd64 trên máy ARM báo "Exec format error" nếu còn dấu này).
TOOL_RUN="$TOOLS/appimagetool-run"
cp "$TOOL" "$TOOL_RUN"
printf '\0\0\0' | dd of="$TOOL_RUN" bs=1 seek=8 count=3 conv=notrunc status=none
chmod +x "$TOOL_RUN"

APPDIR="$(mktemp -d)/Sano.AppDir"
trap 'rm -rf "$(dirname "$APPDIR")"' EXIT
mkdir -p "$APPDIR/usr/bin" "$APPDIR/usr/share/applications" "$APPDIR/usr/share/icons/hicolor/1024x1024/apps"

install -m 0755 "$BIN" "$APPDIR/usr/bin/$APP_NAME"
cp "$DESKTOP_DIR/build/appicon.png" "$APPDIR/sano.png"
cp "$DESKTOP_DIR/build/appicon.png" "$APPDIR/usr/share/icons/hicolor/1024x1024/apps/sano.png"
ln -s sano.png "$APPDIR/.DirIcon"

cat >"$APPDIR/sano.desktop" <<'EOF'
[Desktop Entry]
Type=Application
Name=Sano
Comment=Biến tài liệu của chính bạn thành sách nói, chạy trên máy bạn
Exec=Sano
Icon=sano
Terminal=false
Categories=AudioVideo;Audio;
EOF
cp "$APPDIR/sano.desktop" "$APPDIR/usr/share/applications/sano.desktop"

cat >"$APPDIR/AppRun" <<'EOF'
#!/bin/sh
HERE="$(dirname "$(readlink -f "$0")")"
exec "$HERE/usr/bin/Sano" "$@"
EOF
chmod +x "$APPDIR/AppRun"

mkdir -p "$OUT_DIR"
OUT="$OUT_DIR/$(release_name "$VERSION" linux "$ARCH" AppImage)"
rm -f "$OUT"
log "appimagetool $APPIMAGETOOL_VERSION → $OUT"
# --appimage-extract-and-run: chạy appimagetool không cần FUSE trên máy build.
ARCH=x86_64 VERSION="$VERSION" "$TOOL_RUN" --appimage-extract-and-run \
  --no-appstream --runtime-file "$RUNTIME" "$APPDIR" "$OUT"
chmod +x "$OUT"
log "Xong: $OUT ($(du -h "$OUT" | awk '{print $1}'))"
