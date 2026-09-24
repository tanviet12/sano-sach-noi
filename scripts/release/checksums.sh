#!/usr/bin/env bash
# Tạo SHA256SUMS cho mọi file phát hành trong một thư mục.
#
#   scripts/release/checksums.sh [thư mục]   # mặc định: desktop/build/bin/release
#
# Kiểm lại: sha256sum -c SHA256SUMS (Linux) / shasum -a 256 -c SHA256SUMS (macOS).
# shellcheck source-path=SCRIPTDIR source=common.sh
source "$(dirname "$0")/common.sh"

DIR="${1:-$OUT_DIR}"
[[ -d "$DIR" ]] || die "không thấy thư mục $DIR"

cd "$DIR" || die "không vào được $DIR"
files=()
while IFS= read -r f; do files+=("$f"); done < <(find . -maxdepth 1 -type f ! -name SHA256SUMS ! -name '.*' | sed 's|^\./||' | LC_ALL=C sort)
[[ ${#files[@]} -gt 0 ]] || die "thư mục $DIR không có file nào"

: >SHA256SUMS
for f in "${files[@]}"; do
  printf '%s  %s\n' "$(sha256_of "$f")" "$f" >>SHA256SUMS
done
cat SHA256SUMS
