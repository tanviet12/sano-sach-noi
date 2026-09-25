#!/usr/bin/env bash
# Dựng scripts/tts/vieneu-project/{pyproject.toml,uv.lock}: pyproject.toml của VieNeu-TTS
# đúng VIENEU_COMMIT + ghi đè trong vieneu-overrides.txt, khoá lại bằng uv đúng
# UV_VERSION. Phần mềm desktop chép hai file này vào mã VieNeu (sau khi kiểm tree
# hash) rồi mới `uv sync --frozen`. Sinh lại cả requirements.txt (file đối chiếu).
#
#   scripts/tts/vieneu-lock.sh          # dựng lại
#   scripts/tts/vieneu-lock.sh --check  # chỉ kiểm file trong repo khớp (CI)
set -euo pipefail
HERE="$(cd "$(dirname "$0")" && pwd)"
# versions.env: KEY=VALUE không dấu nháy (bash 3.2 của macOS không source được <(...)).
eval "$(grep -E '^[A-Z0-9_]+=[^ ]*$' "$HERE/versions.env")"
command -v uv >/dev/null || { echo "LỖI: cần uv $UV_VERSION" >&2; exit 1; }
[[ "$(uv --version | awk '{print $2}')" == "$UV_VERSION" ]] || echo "Cảnh báo: uv $(uv --version | awk '{print $2}') khác UV_VERSION=$UV_VERSION" >&2

WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
curl -fsSL --retry 3 "$VIENEU_REPO/archive/$VIENEU_COMMIT.tar.gz" | tar xz -C "$WORK" --strip-components=1

python3 - "$WORK/pyproject.toml" "$HERE/vieneu-overrides.txt" "$VIENEU_COMMIT" <<'PY'
import sys
path, ov, commit = sys.argv[1:]
s = open(path, encoding="utf-8").read()
lines = [l.strip() for l in open(ov, encoding="utf-8") if l.strip() and not l.lstrip().startswith("#")]
anchor = "override-dependencies = ["
i = s.index(anchor) + len(anchor)
extra = "\n    # Sano (scripts/tts/vieneu-overrides.txt):\n" + "".join(f'    "{l}",\n' for l in lines)
s = s[:i] + extra.rstrip("\n") + s[i:]
s = f"# Sano: pyproject.toml của VieNeu-TTS commit {commit} + ghi đè thư viện (scripts/tts/vieneu-lock.sh)\n" + s
open(path, "w", encoding="utf-8").write(s)
PY
(cd "$WORK" && UV_NO_CONFIG=1 uv lock --quiet)

REQ="$WORK/requirements.txt"
{
  cat <<'HDR'
# Thư viện Python phần đọc giọng (VieNeu-TTS) Sano cài, mọi hệ điều hành.
# SINH TỰ ĐỘNG bằng scripts/tts/vieneu-lock.sh từ scripts/tts/vieneu-project/uv.lock
# (uv.lock của VieNeu-TTS tại VIENEU_COMMIT + ghi đè trong vieneu-overrides.txt:
# bỏ giao diện web gradio, bản vá bảo mật). Cách cài: phần mềm desktop chép
# scripts/tts/vieneu-project/ vào mã VieNeu rồi `uv sync --frozen --no-dev`. File này chỉ
# để đọc/đối chiếu; CI kiểm file này khớp (vieneu-lock.sh --check).
HDR
  (cd "$WORK" && uv export --frozen --no-dev --no-hashes --no-emit-project --no-annotate --quiet) | grep -v "^#" | grep -v " ; sys_platform == 'never'$"
} >"$REQ"

if [[ "${1:-}" == "--check" ]]; then
  ok=true
  for f in pyproject.toml uv.lock; do
    cmp -s "$WORK/$f" "$HERE/vieneu-project/$f" || { echo "LỖI: scripts/tts/vieneu-project/$f lệch — chạy scripts/tts/vieneu-lock.sh" >&2; ok=false; }
  done
  cmp -s "$REQ" "$HERE/requirements.txt" || { echo "LỖI: scripts/tts/requirements.txt lệch — chạy scripts/tts/vieneu-lock.sh" >&2; ok=false; }
  $ok && echo "Khớp: scripts/tts/vieneu-project/ và requirements.txt"
  $ok
else
  mkdir -p "$HERE/vieneu-project"
  cp "$WORK/pyproject.toml" "$WORK/uv.lock" "$HERE/vieneu-project/"
  cp "$REQ" "$HERE/requirements.txt"
  echo "Đã dựng scripts/tts/vieneu-project/ ($(grep -c '==' "$HERE/requirements.txt") gói cài) và requirements.txt"
fi
