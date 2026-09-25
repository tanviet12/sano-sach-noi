#!/usr/bin/env bash
# Chạy Sano AppImage thật trên màn hình ảo (Xvfb), bấm chuột bằng xdotool, chụp
# màn hình từng bước. Bộ đọc giả (SANO_TTS_PYTHON/SANO_TTS_SCRIPTS) + điều khoản
# đã đồng ý → vào thẳng thư viện có sách mẫu.
#
#   scripts/ci/linux-gui-smoke.sh play   <Sano.AppImage> <thư mục ảnh>
#   scripts/ci/linux-gui-smoke.sh update <Sano-cũ.AppImage> <thư mục ảnh> <sha256 bản mới>
set -euo pipefail
MODE="$1" APP="$2" OUT="$3" WANT="${4:-}"
mkdir -p "$OUT"
export APPIMAGE_EXTRACT_AND_RUN=1 NO_CLEANUP=1
export SANO_DATA_DIR="$RUNNER_TEMP/sano-data-$MODE" HOME="$RUNNER_TEMP/home-$MODE"
mkdir -p "$SANO_DATA_DIR" "$HOME" "$RUNNER_TEMP/fake-tts"
printf '{"acceptedVersion": 99, "acceptedAt": "2026-09-25T00:00:00Z"}\n' >"$SANO_DATA_DIR/terms.json"
printf 'import sys\nprint("Đủ mô hình (giả)")\nsys.exit(0)\n' >"$RUNNER_TEMP/fake-tts/models.py"
SANO_TTS_PYTHON="$(command -v python3)"
export SANO_TTS_PYTHON SANO_TTS_SCRIPTS="$RUNNER_TEMP/fake-tts"

n=0
shot() { n=$((n + 1)); import -window root "$OUT/$(printf '%02d' $n)-$1.png"; echo "ảnh $n: $1"; }
click() { xdotool mousemove "$1" "$2" click 1; sleep "${3:-2}"; }

run_app() { "$1" >"$OUT/app-$MODE.log" 2>&1 & echo $!; }

if [[ "$MODE" == play ]]; then
  pid=$(run_app "$APP"); sleep 12; shot thu-vien
  xdotool search --name '^Sano$' getwindowgeometry 2>/dev/null | tee "$OUT/cua-so.txt" || true
  click 324 290 3; shot trinh-phat
  click 516 505 5; shot sau-bam-phat
  sleep 3; shot sau-8-giay
  kill "$pid" 2>/dev/null || true
else
  mkdir -p "$HOME/Apps"; cp "$APP" "$HOME/Apps/Sano.AppImage"; chmod +x "$HOME/Apps/Sano.AppImage"
  before=$(sha256sum "$HOME/Apps/Sano.AppImage" | cut -d' ' -f1)
  pid=$(run_app "$HOME/Apps/Sano.AppImage"); sleep 15; shot co-ban-moi
  click 100 646 3; shot hop-cap-nhat
  click 687 490 30; shot tai-xong
  click 681 471 3; shot bam-khoi-dong-lai
  sleep 20; shot sau-mo-lai
  after=$(sha256sum "$HOME/Apps/Sano.AppImage" | cut -d' ' -f1)
  echo "trước: $before"; echo "sau:   $after"; echo "cần:   $WANT"
  pgrep -af Sano || true
  if [[ "$after" == "$WANT" ]]; then echo "ĐẠT: AppImage đã thay bằng bản mới"; else echo "LỖI: AppImage chưa đổi"; exit 1; fi
fi
