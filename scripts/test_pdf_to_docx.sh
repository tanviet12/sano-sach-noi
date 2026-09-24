#!/usr/bin/env bash
# Test pre-process PDF -> DOCX -> zip sách nói (cmd/sano-docx2tts, bước tiền xử lý PDF).
#
# Các bước:
#   1. Sinh PDF mẫu có outline (bookmark) bằng reportlab.
#   2. Chạy pdf_to_docx.py -> /tmp/pdf2docx-test.docx.
#   3. Verify docx hợp lệ: có word/document.xml + có w:pStyle Heading.
#   4. Smoke pipeline: go run ./cmd/sano-docx2tts (--tts-mode stub) -> zip.
#   5. Verify zip valid + có manifest.json + chapters.json.
#
# Yêu cầu: venv scripts/.venv-pdf2docx đã cài requirements.txt (xem README dưới).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

VENV="$ROOT/scripts/.venv-pdf2docx"
PY="$VENV/bin/python"
TMP="$(mktemp -d /tmp/pdf2docx-test.XXXXXX)"
PDF="$TMP/sample.pdf"
DOCX="$TMP/out.docx"
BUILD="$TMP/build"
ZIP="$TMP/from-pdf.zip"

cleanup() { rm -rf "$TMP"; }
trap cleanup EXIT

fail() { echo "✗ FAIL: $*" >&2; exit 1; }

# ── Pre-flight ──────────────────────────────────────────────────────────────
if [ ! -x "$PY" ]; then
  echo "Chưa có venv. Tạo + cài lib:"
  echo "  python3 -m venv scripts/.venv-pdf2docx"
  echo "  scripts/.venv-pdf2docx/bin/pip install -r scripts/requirements.txt"
  fail "thiếu venv $VENV"
fi
"$PY" -c "import pypdf, docx, reportlab" 2>/dev/null || \
  fail "venv thiếu lib (pip install -r scripts/requirements.txt)"

# ── Bước 1: sinh PDF mẫu có outline ──────────────────────────────────────────
echo "[1/5] Sinh PDF mẫu có outline..."
"$PY" - "$PDF" <<'PYEOF'
import sys
from reportlab.pdfgen import canvas
from reportlab.lib.pagesizes import A4

pdf_path = sys.argv[1]
c = canvas.Canvas(pdf_path, pagesize=A4)
W, H = A4

def page(title, body_lines, key, level, parent_closed=False):
    c.bookmarkPage(key)
    c.addOutlineEntry(title, key, level=level, closed=parent_closed)
    c.setFont("Helvetica-Bold", 16)
    c.drawString(72, H - 80, title)
    c.setFont("Helvetica", 12)
    y = H - 120
    for line in body_lines:
        c.drawString(72, y, line)
        y -= 18
    c.showPage()

# Chương 1 (level 0) + 2 tiểu mục (level 1) — mỗi cái 1 trang
page("Chuong 1: Nen tang", ["Moi mo hinh bat dau tu bai toan thuc te."], "c1", 0)
page("Tieu muc 1.1: Khoi dau", ["Hieu ro van de truoc khi thiet ke giai phap."], "c1s1", 1)
page("Tieu muc 1.2: Khach hang", ["Dat khach hang vao trung tam moi quyet dinh."], "c1s2", 1)
# Chương 2 (level 0) + 1 tiểu mục
page("Chuong 2: Dong tien", ["Dong tien la mach mau cua doanh nghiep."], "c2", 0)
page("Tieu muc 2.1: Quan tri", ["Co lai tren so sach van co the thieu tien mat."], "c2s1", 1)

c.save()
print("  PDF:", pdf_path)
PYEOF
[ -s "$PDF" ] || fail "không sinh được PDF mẫu"

# ── Bước 2: convert PDF -> DOCX ──────────────────────────────────────────────
echo "[2/5] Convert PDF -> DOCX..."
"$PY" scripts/pdf_to_docx.py "$PDF" "$DOCX"
[ -s "$DOCX" ] || fail "không sinh được DOCX"

# ── Bước 3: verify docx hợp lệ ───────────────────────────────────────────────
# Lưu ý: capture output ra biến trước rồi mới grep — tránh `grep -q` đóng pipe
# sớm gây SIGPIPE cho unzip (set -o pipefail sẽ coi đó là lỗi).
echo "[3/5] Verify DOCX..."
DOCX_LIST="$(unzip -l "$DOCX")"
echo "$DOCX_LIST" | grep -q "word/document.xml" || fail "DOCX thiếu word/document.xml"
DOC_XML="$(unzip -p "$DOCX" word/document.xml)"
echo "$DOC_XML" | grep -q 'w:pStyle' || fail "DOCX không có w:pStyle (heading)"
echo "$DOC_XML" | grep -qi 'Heading' || fail "DOCX không có style Heading"
echo "  DOCX có Heading style ✓"

# ── Bước 4: smoke pipeline docx -> zip ───────────────────────────────────────
echo "[4/5] Smoke pipeline sano-docx2tts (stub)..."
go run ./cmd/sano-docx2tts \
  --input "$DOCX" \
  --output-dir "$BUILD" \
  --tts-mode stub \
  --output-zip "$ZIP" \
  --author "Test Author" >/dev/null
[ -s "$ZIP" ] || fail "không sinh được zip"

# ── Bước 5: verify zip ───────────────────────────────────────────────────────
echo "[5/5] Verify zip..."
unzip -t "$ZIP" >/dev/null || fail "zip hỏng"
ZIP_LIST="$(unzip -l "$ZIP")"
echo "$ZIP_LIST" | grep -q "manifest.json" || fail "zip thiếu manifest.json"
echo "$ZIP_LIST" | grep -q "chapters.json" || fail "zip thiếu chapters.json"
CHAPTERS_JSON="$(unzip -p "$ZIP" chapters.json)"
CHN=$(echo "$CHAPTERS_JSON" | grep -o '"title"' | wc -l | tr -d ' ')
echo "  zip OK — $(echo "$ZIP_LIST" | tail -1 | awk '{print $2}') entry, chapters.json có $CHN tiêu đề"

echo ""
echo "✓ PASS — PDF -> DOCX -> zip chạy hết pipeline."
