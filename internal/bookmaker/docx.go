package bookmaker

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"
	"unicode"
)

// SectionImage — 1 ảnh nhúng được tham chiếu trong 1 tiểu mục.
type SectionImage struct {
	RelID string // r:embed / r:id trong document.xml
	// Name — tên file do Sano đặt (img%03d<đuôi>, theo thứ tự gặp trong sách),
	// KHÔNG lấy tên trong file Word: tên đó do người soạn tài liệu quyết định,
	// có thể chứa ..\ để ghi ra ngoài thư mục sách.
	Name string
	Data []byte // nội dung file ảnh
}

// Section — đơn vị phát: tiêu đề + văn bản gốc + ảnh kèm.
type Section struct {
	Title  string
	Text   string // văn bản gốc (giữ nguyên lời sách)
	Images []SectionImage
	// Stem — mã gốc "ch%02d-sec%02d" theo thứ tự trong file Word, gán lúc nạp và
	// không đổi khi bỏ bớt tiểu mục (dùng cho DropStems, ReadingEdits, nghe thử).
	Stem string
}

// Chapter — chương chứa các tiểu mục (Heading 2 dưới 1 Heading 1).
type Chapter struct {
	Title    string
	Sections []Section
}

// Book — cấu trúc nhiều cấp trích từ docx.
type Book struct {
	Title    string
	Chapters []Chapter
	Stats    DocStats // số liệu để cảnh báo lúc nạp
}

// DocStats — những thứ bản đọc không truyền tải được hoặc có thể sai cấu trúc.
type DocStats struct {
	Images        int      // hình nhúng (chưa có mô tả → người nghe không biết nội dung)
	SkippedImages int      // hình bỏ qua vì quá lớn sau khi giải nén (chống file "bom nén")
	DroppedImages int      // tham chiếu hình vượt maxImageRefs: bỏ qua
	Tables        int      // bảng (đọc phẳng từng ô, mất cấu trúc hàng/cột)
	FakeHeadings  []string // đoạn in đậm / chữ to nhưng không dùng style Heading
}

// docxPara — 1 đoạn văn đã trích từ word/document.xml.
type docxPara struct {
	style     string   // pStyle val (vd "Heading1", "Title", "")
	text      string   // text gộp các run
	imageRels []string // r:embed / r:id của ảnh trong đoạn
	textRunes int      // số ký tự chữ (không tính khoảng trắng)
	boldRunes int      // số ký tự chữ nằm trong run in đậm
	maxSz     int      // cỡ chữ lớn nhất trong đoạn (nửa point, w:sz); 0 = không ghi
}

// ParseDocx mở file .docx (zip OOXML) → Book nhiều cấp.
//
// Quy ước cấu trúc:
//   - pStyle "Title"              → tiêu đề sách
//   - cấp heading NHỎ NHẤT có mặt → chương (thường Heading1; nếu docx chỉ có
//     Heading2/Heading3 thì Heading2 đóng vai chương)
//   - các cấp heading lớn hơn     → tiểu mục trong chương
//   - đoạn thường                 → văn bản tiểu mục
//
// Ảnh nhúng (<a:blip r:embed> / <v:imagedata r:id>) được trích từ word/media,
// gắn vào tiểu mục chứa nó để bước sau sinh image_description.
func ParseDocx(filePath string) (*Book, error) {
	if err := checkNotProtected(filePath); err != nil {
		return nil, err
	}
	zr, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("mở docx %q: %w", filePath, err)
	}
	defer func() { _ = zr.Close() }()

	docXML, err := readZipFile(&zr.Reader, "word/document.xml", maxXMLBytes)
	if errors.Is(err, errZipEntryTooLarge) {
		return nil, fmt.Errorf("nội dung file Word quá lớn sau khi giải nén (word/document.xml vượt %d MB) — file có thể hỏng hoặc bị làm độc", maxXMLBytes>>20)
	}
	if err != nil {
		return nil, fmt.Errorf("đọc word/document.xml: %w", err)
	}

	paras, tables, err := parseDocumentXML(docXML)
	if err != nil {
		return nil, fmt.Errorf("phân tích document.xml: %w", err)
	}

	relTargets := parseImageRels(&zr.Reader) // relID -> zip path (vd word/media/image1.png)

	book := buildBook(paras, relTargets, &zr.Reader)
	assignStems(book)
	book.Stats.Tables = tables
	if len(book.Chapters) == 0 {
		return nil, fmt.Errorf("docx không có nội dung nào (không Heading/đoạn văn)")
	}
	return book, nil
}

// parseDocumentXML duyệt token document.xml → danh sách đoạn theo thứ tự + số
// bảng. Dùng decoder streaming để xử lý ảnh nhúng lồng sâu tùy ý. Ghi nhận in
// đậm / cỡ chữ theo run (w:r/w:rPr) để phát hiện tiêu đề gõ tay.
func parseDocumentXML(data []byte) ([]docxPara, int, error) {
	dec := xml.NewDecoder(bytes.NewReader(data))
	var paras []docxPara
	idx := -1 // chỉ số đoạn hiện tại
	inText := false
	inPPr := false // trong w:pPr: rPr ở đây là định dạng dấu đoạn, không phải chữ
	runBold, runSz := false, 0
	tables := 0
	// text đoạn hiện tại gom bằng Builder (ghép += chậm theo bình phương với
	// đoạn rất dài); ghi vào paras[idx].text khi hết đoạn hoặc sang đoạn lồng.
	var text strings.Builder
	flush := func() {
		if idx >= 0 {
			paras[idx].text += text.String()
		}
		text.Reset()
	}
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, 0, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "tbl":
				tables++
			case "pPr":
				inPPr = true
			case "r":
				runBold, runSz = false, 0
			case "b":
				if !inPPr {
					runBold = onOffVal(attrVal(t, "val"))
				}
			case "sz":
				if !inPPr {
					runSz, _ = strconv.Atoi(attrVal(t, "val"))
				}
			case "p":
				flush()
				paras = append(paras, docxPara{})
				idx = len(paras) - 1
			case "pStyle":
				if idx >= 0 {
					paras[idx].style = attrVal(t, "val")
				}
			case "t":
				inText = true
			case "tab", "br", "cr":
				if idx >= 0 {
					text.WriteByte(' ')
				}
			case "blip":
				if idx >= 0 {
					if id := attrVal(t, "embed"); id != "" {
						paras[idx].imageRels = append(paras[idx].imageRels, id)
					}
				}
			case "imagedata":
				if idx >= 0 {
					if id := attrVal(t, "id"); id != "" {
						paras[idx].imageRels = append(paras[idx].imageRels, id)
					}
				}
			}
		case xml.CharData:
			if inText && idx >= 0 {
				text.Write(t)
				n := countLetters(string(t))
				paras[idx].textRunes += n
				if runBold {
					paras[idx].boldRunes += n
				}
				if n > 0 && runSz > paras[idx].maxSz {
					paras[idx].maxSz = runSz
				}
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "t":
				inText = false
			case "pPr":
				inPPr = false
			case "p":
				flush()
				idx = -1
			}
		}
	}
	flush()
	return paras, tables, nil
}

// onOffVal đọc thuộc tính bật/tắt OOXML (w:b không có val = bật).
func onOffVal(v string) bool {
	switch strings.ToLower(v) {
	case "0", "false", "off":
		return false
	}
	return true
}

// countLetters đếm ký tự không phải khoảng trắng.
func countLetters(s string) int {
	n := 0
	for _, r := range s {
		if !unicode.IsSpace(r) {
			n++
		}
	}
	return n
}

// buildBook ráp các đoạn thành cấu trúc Book nhiều cấp.
//
// Cấp heading nhỏ nhất xuất hiện trong tài liệu đóng vai chương; các cấp lớn
// hơn là tiểu mục. Nhờ vậy docx export với Heading2/Heading3 (không có Heading1)
// vẫn lên đúng chương/tiểu mục, đồng thời docx chuẩn Heading1/Heading2 giữ
// nguyên hành vi cũ (base = 1).
func buildBook(paras []docxPara, relTargets map[string]string, zr *zip.Reader) *Book {
	book := &Book{}
	var curChapter *Chapter
	var curSection *Section
	imgNames := map[string]string{} // zip path → tên Sano đặt (ảnh dùng lại giữ 1 tên)
	imgData := map[string][]byte{}  // zip path → dữ liệu đã đọc (nil = bỏ qua / lỗi)
	var imgTotal int64              // tổng dung lượng ảnh đã đọc (chặn "bom nén")
	imgRefs := 0                    // số tham chiếu ảnh đã gắn vào tiểu mục (cả cuốn)
	// secImgs — ảnh (zip path) đã gắn vào tiểu mục hiện tại: một ảnh chỉ gắn
	// một lần mỗi tiểu mục dù được tham chiếu lặp lại. Đặt lại khi sang tiểu mục mới.
	secImgs := map[string]bool{}
	// loadImage đọc ảnh một lần cho mỗi zip path, trong giới hạn từng ảnh và tổng.
	loadImage := func(zipPath string) ([]byte, bool) {
		if data, seen := imgData[zipPath]; seen {
			return data, data != nil
		}
		limit := min(maxImageBytes, maxTotalImageBytes-imgTotal)
		data, err := readZipFile(zr, zipPath, limit)
		if errors.Is(err, errZipEntryTooLarge) {
			book.Stats.SkippedImages++
		}
		if err != nil {
			data = nil
		}
		imgData[zipPath] = data
		imgTotal += int64(len(data))
		return data, data != nil
	}
	base := minHeadingLevel(paras)
	bodySz := bodyFontSize(paras)

	// ensureChapter trả chương hiện tại, tạo chương ngầm "Nội dung" nếu chưa có.
	ensureChapter := func() *Chapter {
		if curChapter == nil {
			book.Chapters = append(book.Chapters, Chapter{Title: "Nội dung"})
			curChapter = &book.Chapters[len(book.Chapters)-1]
		}
		return curChapter
	}
	// ensureSection trả tiểu mục hiện tại, tạo ngầm theo tiêu đề chương nếu chưa có.
	ensureSection := func() *Section {
		if curSection == nil {
			ch := ensureChapter()
			ch.Sections = append(ch.Sections, Section{Title: ch.Title})
			curSection = &ch.Sections[len(ch.Sections)-1]
			clear(secImgs)
		}
		return curSection
	}

	for _, p := range paras {
		text := strings.TrimSpace(collapseSpaces(p.text))

		// Caption ảnh (ImageCaption / CaptionedFigure / Caption…) KHÔNG đưa vào
		// lời đọc: người nghe không nhìn màn hình, nhãn "Hình N — ..." vô nghĩa
		// khi nghe. Vẫn giữ ảnh của đoạn (nếu có) để sinh image_description.
		if isCaptionStyle(p.style) {
			text = ""
		}

		switch {
		case isTitleStyle(p.style):
			if text != "" {
				book.Title = text
			}
			continue
		case base > 0 && headingLevel(p.style) == base:
			book.Chapters = append(book.Chapters, Chapter{Title: text})
			curChapter = &book.Chapters[len(book.Chapters)-1]
			curSection = nil
			clear(secImgs)
			continue
		case base > 0 && headingLevel(p.style) > base:
			ch := ensureChapter()
			ch.Sections = append(ch.Sections, Section{Title: text})
			curSection = &ch.Sections[len(ch.Sections)-1]
			clear(secImgs)
			continue
		}

		// Đoạn thường: gắn text + ảnh vào tiểu mục hiện tại.
		if text == "" && len(p.imageRels) == 0 {
			continue
		}
		if !isCaptionStyle(p.style) && looksLikeFakeHeading(p, text, bodySz) {
			book.Stats.FakeHeadings = append(book.Stats.FakeHeadings, text)
		}
		sec := ensureSection()
		if text != "" {
			if sec.Text != "" {
				// Dòng trắng (\n\n) giữa các đoạn → TTS có nghỉ tương xứng giữa đoạn.
				sec.Text += "\n\n"
			}
			sec.Text += text
		}
		for _, rel := range p.imageRels {
			zipPath, ok := relTargets[rel]
			if !ok || secImgs[zipPath] {
				continue
			}
			if imgRefs >= maxImageRefs {
				book.Stats.DroppedImages++
				continue
			}
			data, ok := loadImage(zipPath)
			if !ok {
				continue
			}
			book.Stats.Images++
			name, ok := imgNames[zipPath]
			if !ok {
				name = safeImageName(len(imgNames)+1, zipPath)
				imgNames[zipPath] = name
			}
			sec.Images = append(sec.Images, SectionImage{
				RelID: rel,
				Name:  name,
				Data:  data,
			})
			secImgs[zipPath] = true
			imgRefs++
		}
		// Cập nhật lại con trỏ vì slice có thể đã realloc khi append section ngầm.
		curSection = lastSectionPtr(curChapter)
	}

	pruneEmptySections(book)
	return book
}

// imageExts — đuôi ảnh giữ lại khi đặt tên; đuôi khác → .bin.
var imageExts = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".bmp": true, ".webp": true,
	".emf": true, ".wmf": true, ".tif": true, ".tiff": true,
}

// safeImageName đặt tên ảnh thứ n: img%03d + đuôi (chữ thường) của tên gốc nếu
// nằm trong imageExts, không thì .bin. Kết quả không thể chứa / \ hay "..".
func safeImageName(n int, src string) string {
	ext := strings.ToLower(path.Ext(strings.ReplaceAll(src, `\`, "/")))
	if !imageExts[ext] {
		ext = ".bin"
	}
	return fmt.Sprintf("img%03d%s", n, ext)
}

// lastSectionPtr trả con trỏ tới tiểu mục cuối của chương (an toàn sau realloc).
func lastSectionPtr(ch *Chapter) *Section {
	if ch == nil || len(ch.Sections) == 0 {
		return nil
	}
	return &ch.Sections[len(ch.Sections)-1]
}

// pruneEmptySections bỏ tiểu mục rỗng (không text, không ảnh) + chương rỗng.
func pruneEmptySections(book *Book) {
	chapters := book.Chapters[:0]
	for _, ch := range book.Chapters {
		secs := ch.Sections[:0]
		for _, s := range ch.Sections {
			if strings.TrimSpace(s.Text) == "" && len(s.Images) == 0 {
				continue
			}
			secs = append(secs, s)
		}
		ch.Sections = secs
		if len(ch.Sections) == 0 {
			continue
		}
		chapters = append(chapters, ch)
	}
	book.Chapters = chapters
}

// ── helpers ──────────────────────────────────────────────────────────────

// relsXML — cấu trúc word/_rels/document.xml.rels.
type relsXML struct {
	Rels []struct {
		ID     string `xml:"Id,attr"`
		Type   string `xml:"Type,attr"`
		Target string `xml:"Target,attr"`
	} `xml:"Relationship"`
}

// parseImageRels trả map relID -> zip path (word/media/...) cho các rel ảnh.
func parseImageRels(zr *zip.Reader) map[string]string {
	out := map[string]string{}
	data, err := readZipFile(zr, "word/_rels/document.xml.rels", maxXMLBytes)
	if err != nil {
		return out
	}
	var rels relsXML
	if err := xml.Unmarshal(data, &rels); err != nil {
		return out
	}
	for _, r := range rels.Rels {
		if !strings.Contains(strings.ToLower(r.Type), "image") {
			continue
		}
		target := r.Target
		if strings.HasPrefix(target, "/") {
			target = strings.TrimPrefix(target, "/")
		} else {
			target = "word/" + target // Target tương đối thư mục word/
		}
		out[r.ID] = path.Clean(target)
	}
	return out
}

// Giới hạn dung lượng khi giải nén file Word: file "bom nén" vài KB có thể bung
// ra nhiều GB. Biến (không phải hằng) để test đặt giới hạn nhỏ.
var (
	maxXMLBytes        int64 = 64 << 20  // document.xml, rels…
	maxImageBytes      int64 = 32 << 20  // mỗi ảnh
	maxTotalImageBytes int64 = 512 << 20 // tổng ảnh một tài liệu
	maxImageRefs             = 5000      // tổng số lần gắn ảnh vào tiểu mục một tài liệu
)

// errZipEntryTooLarge — entry bung ra vượt giới hạn cho phép.
var errZipEntryTooLarge = errors.New("vượt giới hạn dung lượng")

// readZipFile đọc 1 entry trong zip theo tên, tối đa max byte. Kiểm kích thước
// khai báo trước khi mở, rồi đọc qua LimitReader phòng header khai sai.
func readZipFile(zr *zip.Reader, name string, max int64) ([]byte, error) {
	for _, f := range zr.File {
		if f.Name != name {
			continue
		}
		if max < 0 || f.UncompressedSize64 > uint64(max) {
			return nil, fmt.Errorf("%q: %w", name, errZipEntryTooLarge)
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		defer func() { _ = rc.Close() }()
		data, err := io.ReadAll(io.LimitReader(rc, max+1))
		if int64(len(data)) > max {
			return nil, fmt.Errorf("%q: %w", name, errZipEntryTooLarge)
		}
		return data, err
	}
	return nil, fmt.Errorf("không tìm thấy %q trong docx", name)
}

func attrVal(se xml.StartElement, local string) string {
	for _, a := range se.Attr {
		if a.Name.Local == local {
			return a.Value
		}
	}
	return ""
}

// minHeadingLevel trả cấp heading nhỏ nhất (>=1) trong tài liệu; 0 nếu không có
// heading nào (toàn đoạn thường / chỉ có Title).
func minHeadingLevel(paras []docxPara) int {
	min := 0
	for _, p := range paras {
		if lvl := headingLevel(p.style); lvl > 0 && (min == 0 || lvl < min) {
			min = lvl
		}
	}
	return min
}

// headingLevel trả cấp heading từ pStyle (Heading1→1, "Heading 2"→2); 0 nếu không.
func headingLevel(style string) int {
	s := strings.ToLower(strings.ReplaceAll(style, " ", ""))
	if !strings.HasPrefix(s, "heading") {
		return 0
	}
	n := strings.TrimPrefix(s, "heading")
	if v, err := strconv.Atoi(n); err == nil && v > 0 {
		return v
	}
	return 0
}

func isTitleStyle(style string) bool {
	return strings.EqualFold(strings.ReplaceAll(style, " ", ""), "title")
}

// isCaptionStyle nhận diện style caption ảnh/bảng (ImageCaption, CaptionedFigure,
// Caption, TableCaption…) — text các đoạn này là nhãn hình, không đọc thành audio.
func isCaptionStyle(style string) bool {
	return strings.Contains(strings.ToLower(style), "caption")
}

// collapseSpaces gộp khoảng trắng/tab/xuống dòng liền nhau thành 1 space.
func collapseSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// assignStems gán mã gốc ch%02d-sec%02d cho mọi tiểu mục theo thứ tự hiện tại.
func assignStems(book *Book) {
	for ci := range book.Chapters {
		for si := range book.Chapters[ci].Sections {
			book.Chapters[ci].Sections[si].Stem = fmt.Sprintf("ch%02d-sec%02d", ci+1, si+1)
		}
	}
}
