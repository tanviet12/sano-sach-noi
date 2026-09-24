#!/usr/bin/env bash
# In nội dung GitHub Release cho một phiên bản: mục tương ứng trong CHANGELOG.md
# (tiêu đề "## 1.2.3", "## v1.2.3" hoặc "## [1.2.3]"), rồi hướng dẫn tải + kiểm.
#
#   scripts/release/notes.sh <version> [owner/repo]
# shellcheck source-path=SCRIPTDIR source=common.sh
source "$(dirname "$0")/common.sh"

VERSION="${1:-}"
REPO="${2:-${GITHUB_REPOSITORY:-OWNER/REPO}}"
check_version "$VERSION"

section=""
if [[ -f "$REPO_ROOT/CHANGELOG.md" ]]; then
  section="$(awk -v v="$VERSION" '
    /^## / {
      if (grab) exit
      h = $0
      sub(/^## +/, "", h); gsub(/[\[\]]/, "", h); sub(/^v/, "", h)
      split(h, parts, /[ \t]/)
      if (parts[1] == v) { grab = 1; next }
    }
    grab { print }
  ' "$REPO_ROOT/CHANGELOG.md")"
fi

if [[ -n "${section//[[:space:]]/}" ]]; then
  printf '%s\n' "$section"
else
  printf 'Bản %s. Chi tiết thay đổi xem CHANGELOG.md.\n' "$VERSION"
fi

cat <<EOF

## Tải bản cài

| Máy | File |
|---|---|
| Windows 10/11 (64-bit) | \`Sano-$VERSION-windows-amd64-setup.exe\` (bộ cài, không cần quyền admin) hoặc \`Sano-$VERSION-windows-amd64-portable.zip\` (giải nén là chạy) |
| macOS 10.13+ (Apple Silicon + Intel) | \`Sano-$VERSION-macos-universal.dmg\` |
| Linux x86_64 (cần WebKitGTK 4.1) | \`Sano-$VERSION-linux-amd64.AppImage\` |

**App chưa ký số.** Lần đầu mở: macOS báo không xác minh được nhà phát triển → Cài đặt hệ thống → Quyền riêng tư & Bảo mật → **Vẫn mở**. Windows SmartScreen → **Thông tin thêm** → **Vẫn chạy**. Linux: \`chmod +x\` rồi chạy.

## Kiểm file tải về

\`\`\`bash
sha256sum -c SHA256SUMS --ignore-missing        # Linux
shasum -a 256 -c SHA256SUMS --ignore-missing    # macOS
gh attestation verify <file> --repo $REPO       # nguồn gốc build (GitHub CLI)
\`\`\`

Windows (PowerShell): \`Get-FileHash .\\<file> -Algorithm SHA256\` rồi so với dòng tương ứng trong \`SHA256SUMS\`.

Mọi file được build trên GitHub Actions từ đúng thẻ \`v$VERSION\`, không build tay.
EOF
