package bookmaker

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"path/filepath"
	"strings"
)

// ParseBook mở file nguồn theo đuôi: .epub → ParseEpub, .pdf → ParsePDF, còn lại → ParseDocx.
func ParseBook(filePath string) (*Book, error) {
	switch {
	case IsEpub(filePath):
		return ParseEpub(filePath)
	case IsPDF(filePath):
		return ParsePDF(filePath)
	}
	return ParseDocx(filePath)
}

// IsPDF — file nguồn là PDF (theo đuôi file).
func IsPDF(filePath string) bool {
	return strings.EqualFold(filepath.Ext(filePath), ".pdf")
}

// IsEpub — file nguồn là sách điện tử EPUB (theo đuôi file).
func IsEpub(filePath string) bool {
	return strings.EqualFold(filepath.Ext(filePath), ".epub")
}

// IsSourceFile — đuôi file Sano nạp được làm sách: .docx, .epub hoặc .pdf.
func IsSourceFile(filePath string) bool {
	return strings.EqualFold(filepath.Ext(filePath), ".docx") || IsEpub(filePath) || IsPDF(filePath)
}

// maxEpubTextBytes — tổng dung lượng các trang nội dung (XHTML) đọc từ một EPUB.
var maxEpubTextBytes int64 = 256 << 20

// ParseEpub mở file .epub (zip, EPUB 2 hoặc 3) → Book nhiều cấp, cùng quy ước
// với ParseDocx để các bước sau (chuẩn hoá chữ, đọc, xuất M4B) dùng chung:
//   - các trang nội dung đọc theo thứ tự spine trong file OPF
//   - h1…h6 trong trang → Heading1…6 (cấp nhỏ nhất có mặt là chương)
//   - trang không có thẻ h nào → lấy tên trong mục lục (nav / NCX) làm tiêu đề
//   - tên sách lấy từ dc:title; thẻ h trùng tên sách ở đầu sách không thành chương
//
// Bỏ qua: trang không có chữ (bìa, trang chỉ có ảnh), trang mục lục nav,
// chú thích cuối trang (epub:type footnote / endnote), số tham chiếu chú thích.
// EPUB có DRM (nội dung mã hoá) bị từ chối như file Word có khoá.
func ParseEpub(filePath string) (*Book, error) {
	zr, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("mở epub %q: %w", filePath, err)
	}
	defer func() { _ = zr.Close() }()
	z := &zr.Reader

	opfPath, err := epubRootfile(z)
	if err != nil {
		return nil, err
	}
	opfData, err := readZipFile(z, opfPath, maxXMLBytes)
	if err != nil {
		return nil, fmt.Errorf("đọc %s: %w", opfPath, err)
	}
	var opf opfXML
	if err := newLooseDecoder(opfData).Decode(&opf); err != nil {
		return nil, fmt.Errorf("phân tích %s: %w", opfPath, err)
	}
	opfDir := path.Dir(opfPath)

	items := map[string]opfItem{} // id → item (href đã ghép thành đường dẫn trong zip)
	for _, it := range opf.Manifest {
		it.Href = zipJoin(opfDir, it.Href)
		items[it.ID] = it
	}
	if err := checkEpubDRM(z, items); err != nil {
		return nil, err
	}

	toc := epubTOC(z, opf, items)
	title := collapseSpaces(strings.TrimSpace(opf.Title))

	var paras []docxPara
	var textBytes int64
	tables := 0
	for _, ref := range opf.Spine.Items {
		it, ok := items[ref.IDRef]
		if !ok || strings.EqualFold(ref.Linear, "no") || hasProperty(it.Properties, "nav") || !isXHTML(it.MediaType) {
			continue
		}
		limit := min(maxXMLBytes, maxEpubTextBytes-textBytes)
		data, err := readZipFile(z, it.Href, limit)
		if errors.Is(err, errZipEntryTooLarge) {
			return nil, fmt.Errorf("nội dung EPUB quá lớn sau khi giải nén (%s) — file có thể hỏng hoặc bị làm độc", it.Href)
		}
		if err != nil {
			continue // trang khai trong spine nhưng thiếu trong zip: bỏ qua
		}
		textBytes += int64(len(data))
		ps, n := parseEpubPage(data, it.Href, toc[it.Href])
		tables += n
		paras = append(paras, ps...)
	}
	markTitleHeadings(paras, title)

	relTargets := map[string]string{} // ảnh EPUB: "rel" chính là đường dẫn trong zip
	for _, p := range paras {
		for _, r := range p.imageRels {
			relTargets[r] = r
		}
	}
	book := buildBook(paras, relTargets, z)
	if book.Title == "" {
		book.Title = title
	}
	assignStems(book)
	book.Stats.Tables = tables
	if len(book.Chapters) == 0 {
		return nil, fmt.Errorf("EPUB không có chữ để đọc (sách chỉ có hình như truyện tranh, hoặc chữ nằm trong ảnh)")
	}
	return book, nil
}

// ── OPF, container ───────────────────────────────────────────────────────

type opfItem struct {
	ID         string `xml:"id,attr"`
	Href       string `xml:"href,attr"`
	MediaType  string `xml:"media-type,attr"`
	Properties string `xml:"properties,attr"`
}

type opfXML struct {
	Title    string    `xml:"metadata>title"`
	Manifest []opfItem `xml:"manifest>item"`
	Spine    struct {
		TOC   string `xml:"toc,attr"` // EPUB 2: id của file NCX
		Items []struct {
			IDRef  string `xml:"idref,attr"`
			Linear string `xml:"linear,attr"`
		} `xml:"itemref"`
	} `xml:"spine"`
}

// epubRootfile đọc META-INF/container.xml → đường dẫn file OPF.
func epubRootfile(z *zip.Reader) (string, error) {
	data, err := readZipFile(z, "META-INF/container.xml", maxXMLBytes)
	if err != nil {
		return "", fmt.Errorf("file không phải EPUB hợp lệ (thiếu META-INF/container.xml)")
	}
	var c struct {
		Rootfiles []struct {
			FullPath  string `xml:"full-path,attr"`
			MediaType string `xml:"media-type,attr"`
		} `xml:"rootfiles>rootfile"`
	}
	if err := newLooseDecoder(data).Decode(&c); err != nil {
		return "", fmt.Errorf("phân tích META-INF/container.xml: %w", err)
	}
	for _, r := range c.Rootfiles {
		if r.FullPath != "" && (r.MediaType == "" || strings.Contains(r.MediaType, "oebps")) {
			return path.Clean(r.FullPath), nil
		}
	}
	return "", fmt.Errorf("file không phải EPUB hợp lệ (container.xml không chỉ tới file OPF)")
}

// checkEpubDRM từ chối EPUB có nội dung bị mã hoá (DRM Adobe, Readium LCP…).
// Font bị "làm rối" (obfuscation) theo chuẩn IDPF/Adobe không phải DRM: cho qua.
func checkEpubDRM(z *zip.Reader, items map[string]opfItem) error {
	for _, f := range z.File {
		if f.Name == "META-INF/rights.xml" || f.Name == "META-INF/license.lcpl" {
			return ErrProtectedFile
		}
	}
	data, err := readZipFile(z, "META-INF/encryption.xml", maxXMLBytes)
	if err != nil {
		return nil
	}
	var enc struct {
		Data []struct {
			Method struct {
				Algorithm string `xml:"Algorithm,attr"`
			} `xml:"EncryptionMethod"`
		} `xml:"EncryptedData"`
	}
	if err := newLooseDecoder(data).Decode(&enc); err != nil {
		return ErrProtectedFile
	}
	for _, d := range enc.Data {
		alg := d.Method.Algorithm
		if alg == "http://www.idpf.org/2008/embedding" || alg == "http://ns.adobe.com/pdf/enc#RC" {
			continue
		}
		return ErrProtectedFile
	}
	return nil
}

// ── mục lục: nav (EPUB 3) hoặc NCX (EPUB 2) ──────────────────────────────

// tocEntry — 1 mục trong mục lục sách: tiêu đề, cấp (1 = ngoài cùng), neo trong trang.
type tocEntry struct {
	Title string
	Level int
	Frag  string // id phần tử trong trang; trống = đầu trang
}

// epubTOC trả map đường dẫn trang → các mục lục trỏ vào trang đó (theo thứ tự).
// Không có mục lục hoặc mục lục hỏng → map rỗng (chỉ dựa vào thẻ h trong trang).
func epubTOC(z *zip.Reader, opf opfXML, items map[string]opfItem) map[string][]tocEntry {
	out := map[string][]tocEntry{}
	add := func(base, href, title string, level int) {
		title = collapseSpaces(strings.TrimSpace(title))
		if href == "" || title == "" {
			return
		}
		file, frag, _ := strings.Cut(href, "#")
		file = zipJoin(path.Dir(base), file)
		out[file] = append(out[file], tocEntry{Title: title, Level: level, Frag: frag})
	}
	for _, it := range items {
		if hasProperty(it.Properties, "nav") {
			if data, err := readZipFile(z, it.Href, maxXMLBytes); err == nil {
				parseNavDoc(data, func(href, title string, level int) { add(it.Href, href, title, level) })
			}
			if len(out) > 0 {
				return out
			}
		}
	}
	ncx, ok := items[opf.Spine.TOC]
	if !ok {
		for _, it := range items {
			if it.MediaType == "application/x-dtbncx+xml" {
				ncx, ok = it, true
				break
			}
		}
	}
	if ok {
		if data, err := readZipFile(z, ncx.Href, maxXMLBytes); err == nil {
			parseNCX(data, func(href, title string, level int) { add(ncx.Href, href, title, level) })
		}
	}
	return out
}

// parseNavDoc duyệt <nav epub:type="toc"> của EPUB 3: mỗi <a href> là một mục,
// cấp theo độ sâu <ol> lồng nhau.
func parseNavDoc(data []byte, add func(href, title string, level int)) {
	dec := newHTMLDecoder(data)
	inTOC, navDepth, olDepth := false, 0, 0
	var href string
	var text strings.Builder
	inA := false
	for {
		tok, err := dec.Token()
		if err != nil {
			return
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "nav":
				navDepth++
				if !inTOC && attrVal(t, "type") == "toc" {
					inTOC = true
					navDepth = 1
				}
			case "ol", "ul":
				if inTOC {
					olDepth++
				}
			case "a":
				if inTOC {
					inA, href = true, attrVal(t, "href")
					text.Reset()
				}
			}
		case xml.CharData:
			if inA {
				text.Write(t)
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "nav":
				navDepth--
				if inTOC && navDepth == 0 {
					return
				}
			case "ol", "ul":
				if inTOC {
					olDepth--
				}
			case "a":
				if inA {
					add(href, text.String(), max(1, olDepth))
					inA = false
				}
			}
		}
	}
}

// parseNCX duyệt navMap của EPUB 2: mỗi navPoint (lồng nhau) là một mục.
func parseNCX(data []byte, add func(href, title string, level int)) {
	dec := newLooseDecoder(data)
	depth := 0
	inText := false
	var label strings.Builder
	pending := false // navPoint hiện tại chưa gặp <content>
	for {
		tok, err := dec.Token()
		if err != nil {
			return
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "navPoint":
				depth++
				label.Reset()
				pending = true
			case "text":
				inText = pending
			case "content":
				if pending {
					add(attrVal(t, "src"), label.String(), depth)
					pending = false
				}
			}
		case xml.CharData:
			if inText {
				label.Write(t)
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "navPoint":
				depth--
			case "text":
				inText = false
			}
		}
	}
}

// ── trang nội dung XHTML ─────────────────────────────────────────────────

// epubBlocks — thẻ khối: mỗi thẻ mở/đóng là ranh giới đoạn văn.
var epubBlocks = map[string]bool{
	"p": true, "div": true, "li": true, "blockquote": true, "pre": true, "section": true,
	"article": true, "dt": true, "dd": true, "td": true, "th": true, "tr": true,
	"figcaption": true, "caption": true, "header": true, "footer": true, "figure": true,
	"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
}

// epubSkip — thẻ bỏ cả nội dung bên trong (không phải lời sách).
var epubSkip = map[string]bool{"head": true, "script": true, "style": true, "nav": true, "rt": true, "rp": true}

// epubNoteTypes — epub:type của chú thích và số tham chiếu chú thích: không đọc.
var epubNoteTypes = []string{"footnote", "endnote", "rearnote", "note", "noteref", "pagebreak"}

// parseEpubPage trích các đoạn của 1 trang XHTML theo thứ tự. toc — mục lục trỏ
// vào trang: dùng làm tiêu đề khi trang không có thẻ h nào. Trang không có chữ
// (bìa, trang chỉ có ảnh) trả rỗng.
func parseEpubPage(data []byte, pagePath string, toc []tocEntry) ([]docxPara, int) {
	dec := newHTMLDecoder(data)
	var paras []docxPara
	var cur *docxPara
	var text strings.Builder
	tables := 0
	skip := 0                   // độ sâu trong vùng bỏ qua (head, script, chú thích…)
	anchors := map[string]int{} // id phần tử → chỉ số đoạn chứa / ngay sau nó

	flush := func() {
		if cur != nil {
			cur.text += text.String()
			cur.textRunes = countLetters(cur.text)
			if strings.TrimSpace(cur.text) != "" || len(cur.imageRels) > 0 {
				paras = append(paras, *cur)
			}
		}
		text.Reset()
		cur = nil
	}
	open := func(style string) {
		flush()
		cur = &docxPara{style: style}
	}
	inBody := false
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			name := strings.ToLower(t.Name.Local)
			if skip > 0 || epubSkip[name] || isNoteElement(t) {
				skip++
				continue
			}
			if name == "body" {
				inBody = true
			}
			if !inBody {
				continue
			}
			switch {
			case name == "table":
				tables++
			case len(name) == 2 && name[0] == 'h' && name[1] >= '1' && name[1] <= '6':
				open("Heading" + name[1:])
			case name == "figcaption" || name == "caption":
				open("Caption")
			case epubBlocks[name]:
				open("")
			case name == "br":
				text.WriteByte(' ')
			case name == "img" || name == "image":
				src := attrVal(t, "src")
				if src == "" {
					src = attrVal(t, "href") // <svg:image xlink:href>
				}
				if src != "" && !strings.HasPrefix(src, "data:") {
					if cur == nil {
						open("")
					}
					cur.imageRels = append(cur.imageRels, zipJoin(path.Dir(pagePath), src))
				}
			}
			// Sau open(): thẻ khối → đoạn mới mở ở chỉ số len(paras); thẻ trong
			// dòng → đoạn đang gom cũng sẽ ở len(paras).
			if id := attrVal(t, "id"); id != "" {
				if _, seen := anchors[id]; !seen {
					anchors[id] = len(paras)
				}
			}
		case xml.CharData:
			if skip > 0 || !inBody {
				continue
			}
			if cur == nil {
				open("")
			}
			text.Write(t)
		case xml.EndElement:
			if skip > 0 {
				skip--
				continue
			}
			if epubBlocks[strings.ToLower(t.Name.Local)] {
				flush()
			}
		}
	}
	flush()

	for _, p := range paras {
		if strings.TrimSpace(p.text) != "" {
			return withTOCHeadings(paras, toc, anchors), tables
		}
	}
	return nil, tables // trang không có chữ: bìa, trang ảnh
}

// withTOCHeadings chèn tiêu đề từ mục lục vào trang không có thẻ h nào: mục trỏ
// đầu trang chèn ở đầu, mục trỏ neo #id chèn trước phần tử đó. Đoạn ngay sau
// trùng chữ với tiêu đề (tiêu đề gõ bằng <p class="...">) thì bỏ để không đọc lặp.
func withTOCHeadings(paras []docxPara, toc []tocEntry, anchors map[string]int) []docxPara {
	if len(toc) == 0 {
		return paras
	}
	for _, p := range paras {
		if headingLevel(p.style) > 0 {
			return paras
		}
	}
	at := map[int][]tocEntry{}
	for _, e := range toc {
		i := 0
		if e.Frag != "" {
			var ok bool
			if i, ok = anchors[e.Frag]; !ok {
				continue
			}
		}
		at[i] = append(at[i], e)
	}
	out := make([]docxPara, 0, len(paras)+len(toc))
	for i := 0; i <= len(paras); i++ {
		for _, e := range at[i] {
			out = append(out, docxPara{style: fmt.Sprintf("Heading%d", min(e.Level, 6)), text: e.Title, textRunes: countLetters(e.Title)})
			if i < len(paras) && sameText(paras[i].text, e.Title) && len(paras[i].imageRels) == 0 {
				paras[i].text = ""
			}
		}
		if i < len(paras) && (strings.TrimSpace(paras[i].text) != "" || len(paras[i].imageRels) > 0) {
			out = append(out, paras[i])
		}
	}
	return out
}

// markTitleHeadings: thẻ h ghi tên sách ở đầu sách (trang tên sách) đổi thành
// style Title, để không thành một chương rỗng nuốt cả cuốn làm tiểu mục.
func markTitleHeadings(paras []docxPara, title string) {
	if title == "" {
		return
	}
	for i := range paras {
		if headingLevel(paras[i].style) == 0 {
			continue
		}
		if !sameText(paras[i].text, title) {
			return
		}
		paras[i].style = "Title"
	}
}

// ── helpers ──────────────────────────────────────────────────────────────

// newLooseDecoder — decoder XML không khắt khe, hiểu thực thể HTML (&nbsp;…).
// Dùng cho file XML của EPUB (OPF, NCX, container): KHÔNG tự đóng thẻ kiểu HTML
// vì OPF có <meta>…</meta> chứa chữ.
func newLooseDecoder(data []byte) *xml.Decoder {
	dec := xml.NewDecoder(bytes.NewReader(data))
	dec.Strict = false
	dec.Entity = xml.HTMLEntity
	dec.CharsetReader = func(_ string, r io.Reader) (io.Reader, error) { return r, nil }
	return dec
}

// newHTMLDecoder — như newLooseDecoder, thêm tự đóng thẻ rỗng HTML (<br>, <img>
// không có />) cho trang nội dung XHTML lỏng của EPUB 2.
func newHTMLDecoder(data []byte) *xml.Decoder {
	dec := newLooseDecoder(data)
	dec.AutoClose = xml.HTMLAutoClose
	return dec
}

// zipJoin ghép href (tương đối, có thể %-mã hoá) với thư mục chứa nó → đường dẫn trong zip.
func zipJoin(dir, href string) string {
	href, _, _ = strings.Cut(href, "#")
	if u, err := url.PathUnescape(href); err == nil {
		href = u
	}
	if strings.HasPrefix(href, "/") {
		return path.Clean(strings.TrimPrefix(href, "/"))
	}
	return strings.TrimPrefix(path.Clean(path.Join(dir, href)), "./")
}

func hasProperty(props, want string) bool {
	for _, p := range strings.Fields(props) {
		if p == want {
			return true
		}
	}
	return false
}

func isXHTML(mediaType string) bool {
	return mediaType == "application/xhtml+xml" || mediaType == "text/html"
}

func isNoteElement(t xml.StartElement) bool {
	for _, a := range t.Attr {
		if a.Name.Local != "type" && a.Name.Local != "role" {
			continue
		}
		for _, v := range strings.Fields(strings.ToLower(a.Value)) {
			v = strings.TrimPrefix(v, "doc-") // role="doc-footnote"
			for _, n := range epubNoteTypes {
				if v == n {
					return true
				}
			}
		}
	}
	return false
}

func sameText(a, b string) bool {
	return strings.EqualFold(collapseSpaces(strings.TrimSpace(a)), collapseSpaces(strings.TrimSpace(b)))
}
