#!/usr/bin/env bash
# Quét các file cài (Sano-*) bằng VirusTotal, in bảng kết quả Markdown.
#
#   VT_API_KEY=... scripts/release/virustotal.sh <thư mục> > virustotal.md
#
# - File đã có trên VirusTotal (cùng SHA256) thì lấy kết quả sẵn, không gửi lại.
# - Chưa có thì gửi lên (gói miễn phí: file ≤ 32 MB), đợi quét xong (tối đa ~15 phút/file).
# - Gói miễn phí giới hạn 4 lượt gọi/phút → mỗi lượt gọi cách nhau 16 giây.
# - Có phần mềm diệt virus báo nhiễm: in cảnh báo GitHub (::warning), KHÔNG dừng —
#   app chưa ký số hay bị báo nhầm; xử lý bằng cách gửi file cho hãng phân tích.
# Lỗi mạng / API thì thoát khác 0 (CI để bước này không chặn phát hành).
# shellcheck source-path=SCRIPTDIR source=common.sh
source "$(dirname "$0")/common.sh"

DIR="${1:-$OUT_DIR}"
[[ -d "$DIR" ]] || die "không thấy thư mục $DIR"
[[ -n "${VT_API_KEY:-}" ]] || die "thiếu VT_API_KEY"
command -v jq >/dev/null || die "thiếu jq"
API="${VT_API:-https://www.virustotal.com/api/v3}"
GUI="https://www.virustotal.com/gui/file"
MAX_UPLOAD=$((32 * 1024 * 1024))
PACE="${VT_PACE:-16}" # giây giữa hai lượt gọi

# Giữ nhịp ở tiến trình chính (vt chạy trong $(...) là tiến trình con, không
# nhớ được lần gọi trước) — gọi pace ngay trước mỗi lượt vt.
last=0
pace() {
  local wait=$((last + PACE - $(date +%s)))
  if ((wait > 0)); then sleep "$wait"; fi
  last=$(date +%s)
}
vt() { # vt <curl args...> — in JSON trả về, lỗi HTTP thì thoát khác 0
  curl -sS --fail-with-body --retry 3 --max-time 300 -H "x-apikey: $VT_API_KEY" "$@"
}

files=()
while IFS= read -r f; do files+=("$f"); done < <(find "$DIR" -maxdepth 1 -type f -name 'Sano-*' | LC_ALL=C sort)
[[ ${#files[@]} -gt 0 ]] || die "thư mục $DIR không có file Sano-*"

rows=()
for f in "${files[@]}"; do
  name="$(basename "$f")"
  sha="$(sha256_of "$f")"
  log "VirusTotal: $name ($sha)" >&2
  stats=""
  pace
  if out="$(vt "$API/files/$sha" 2>/dev/null)"; then
    stats="$(jq -c '.data.attributes.last_analysis_stats' <<<"$out")"
    # Có bản ghi nhưng chưa quét xong lần nào → đợi như file mới gửi.
    [[ "$(jq '(. // {}) | [.[]] | add // 0' <<<"$stats")" -gt 0 ]] || stats=""
  fi
  if [[ -z "$stats" ]]; then
    size=$(wc -c <"$f" | tr -d ' ')
    ((size <= MAX_UPLOAD)) || die "$name lớn hơn 32 MB (gói miễn phí không gửi được)"
    pace
    id="$(vt -F "file=@$f" "$API/files" | jq -r '.data.id')"
    [[ -n "$id" && "$id" != null ]] || die "gửi $name lên VirusTotal không trả mã phân tích"
    for _ in $(seq 1 55); do # 55 × 16 giây ≈ 15 phút
      pace
      out="$(vt "$API/analyses/$id")"
      if [[ "$(jq -r '.data.attributes.status' <<<"$out")" == completed ]]; then
        stats="$(jq -c '.data.attributes.stats' <<<"$out")"
        break
      fi
    done
    [[ -n "$stats" ]] || die "VirusTotal quét $name quá lâu"
  fi
  mal=$(jq '.malicious // 0' <<<"$stats")
  sus=$(jq '.suspicious // 0' <<<"$stats")
  total=$(jq '(. // {}) | [.[]] | add // 0' <<<"$stats")
  if ((mal + sus > 0)); then
    echo "::warning::VirusTotal: $name bị $mal phần mềm báo độc hại, $sus báo đáng ngờ — xem $GUI/$sha" >&2
  fi
  rows+=("| \`$name\` | $mal độc hại · $sus đáng ngờ / $total | [xem]($GUI/$sha) |")
done

printf '## Quét VirusTotal\n\n| File | Kết quả | Chi tiết |\n|---|---|---|\n'
printf '%s\n' "${rows[@]}"
printf '\nApp chưa ký số nên đôi khi vài phần mềm diệt virus báo nhầm; mã nguồn và cách build công khai ở repo này.\n'
