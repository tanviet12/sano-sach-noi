package bookmaker

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ErrPDFNoText — PDF không có lớp chữ (bản scan, ảnh chụp).
var ErrPDFNoText = errors.New("PDF không có chữ để đọc (bản scan hoặc ảnh chụp). Sano chưa đọc được loại này: cần PDF có chữ bôi đen chọn được, hoặc file Word")

// ErrPDFGarbled — chữ trong PDF hỏng gần hết (phông cũ, phông không có bảng mã).
var ErrPDFGarbled = errors.New("chữ trong PDF bị lỗi phông gần hết (thường do phông cũ TCVN3, VNI hoặc phông không có bảng mã Unicode), bộ đọc sẽ đọc sai. Dùng file Word gốc, hoặc xuất lại PDF bằng phông Unicode")

// Ngưỡng đánh giá chất lượng chữ PDF (tỉ lệ ký tự lỗi trên số chữ cái).
const (
	pdfGarbleNote   = 0.01 // từ 1%: cảnh báo
	pdfGarbleSevere = 0.10 // từ 10%: cảnh báo nặng
	pdfGarbleRefuse = 0.60 // từ 60%: từ chối (nghe không ra gì)
	pdfMinLetters   = 100  // ít hơn: coi như không có chữ
	pdfPageLetters  = 20   // trang ít hơn số chữ này: trang không có chữ
)

// ParsePDF mở file .pdf có lớp chữ → Book nhiều cấp, cùng quy ước với ParseDocx:
//   - chương, mục lấy từ bookmark (mục lục bên lề PDF); không có bookmark thì
//     đoán theo cỡ chữ tiêu đề (cỡ lớn nhất = chương)
//   - bỏ đầu trang, chân trang, số trang lặp lại; ghép dòng thành đoạn, nối từ
//     bị ngắt bằng gạch nối, nối đoạn bị cắt ngang giữa hai trang
//   - PDF có mật khẩu → ErrProtectedFile; không có chữ → ErrPDFNoText; chữ lỗi
//     phông gần hết → ErrPDFGarbled
//
// Những gì ảnh hưởng chất lượng nghe (thiếu bookmark, trang ảnh, lỗi phông,
// nhiều cột) ghi vào Stats.Notes để cảnh báo người dùng lúc nạp.
func ParsePDF(filePath string) (*Book, error) {
	doc, err := readPDF(filePath)
	if err != nil {
		return nil, err
	}
	return bookFromPDF(doc)
}

// bookFromPDF dựng Book từ nội dung thô (tách riêng để test không cần PDFium).
func bookFromPDF(src *pdfDoc) (*Book, error) {
	doc := src.clone() // bản trong cache dùng lại nhiều lần: không sửa trực tiếp
	q := pdfQuality(doc)
	if q.letters < pdfMinLetters {
		return nil, ErrPDFNoText
	}
	if q.ratio() >= pdfGarbleRefuse {
		return nil, ErrPDFGarbled
	}

	stripRunningLines(doc)
	paras := pdfParagraphs(doc)
	var notes []LoadNote
	usedOutline := len(usableOutline(doc)) >= 2
	if usedOutline {
		paras = applyPDFOutline(paras, usableOutline(doc))
	} else {
		n := markPDFHeadings(paras)
		if n == 0 {
			notes = append(notes, LoadNote{Text: "PDF không có mục lục (bookmark) và không nhận ra tiêu đề nào: cả file thành một chương. Muốn chia chương, dùng file Word, hoặc xuất PDF có bookmark"})
		} else {
			notes = append(notes, LoadNote{Text: "PDF không có mục lục (bookmark): Sano đoán chương, mục theo cỡ chữ tiêu đề. Xem kỹ ở bước Mục lục, bỏ tick mục thừa"})
		}
	}
	title := pdfTitle(doc.Title)
	docParas := make([]docxPara, 0, len(paras))
	for _, p := range paras {
		docParas = append(docParas, p.para)
	}
	markTitleHeadings(docParas, title)
	book := buildBook(docParas, nil, nil)
	if book.Title == "" {
		book.Title = title
	}
	assignStems(book)
	if len(book.Chapters) == 0 {
		return nil, ErrPDFNoText
	}

	images := 0
	for _, p := range doc.Pages {
		images += p.Images
	}
	book.Stats.Images = images
	if r := q.ratio(); r >= pdfGarbleNote {
		pct := int(math.Round(r * 100))
		if pct < 1 {
			pct = 1
		}
		notes = append(notes, LoadNote{
			Severe: r >= pdfGarbleSevere,
			Text: fmt.Sprintf("Khoảng %d%% chữ bị lỗi phông (thường do phông cũ TCVN3, VNI hoặc phông không có bảng mã Unicode): bộ đọc sẽ đọc sai những chỗ này%s. Nên dùng file Word gốc, hoặc xuất lại PDF bằng phông Unicode",
				pct, pageList(q.badPages)),
		})
	}
	if n := len(q.emptyPages); n > 0 {
		notes = append(notes, LoadNote{
			Severe: n*2 >= len(doc.Pages),
			Text:   fmt.Sprintf("%d/%d trang không có chữ (trang ảnh hoặc bản scan)%s: nội dung các trang này sẽ không được đọc", n, len(doc.Pages), pageList(q.emptyPages)),
		})
	}
	if pages := multiColumnPages(doc); len(pages) > 0 {
		notes = append(notes, LoadNote{Text: fmt.Sprintf("%d trang dàn nhiều cột%s: thứ tự đọc có thể bị đảo ở vài chỗ, nên nghe thử những trang này", len(pages), pageList(pages))})
	}
	book.Stats.Notes = notes
	return book, nil
}

func (d *pdfDoc) clone() *pdfDoc {
	c := *d
	c.Pages = make([]pdfPage, len(d.Pages))
	for i, p := range d.Pages {
		c.Pages[i] = p
		c.Pages[i].Lines = append([]pdfLine(nil), p.Lines...)
	}
	return &c
}

// ── chất lượng chữ ───────────────────────────────────────────────────────

type pdfQualityStats struct {
	letters, bad int
	badPages     []int // trang (1-based) có từ 5% ký tự lỗi
	emptyPages   []int // trang (1-based) không có chữ
}

func (q pdfQualityStats) ratio() float64 {
	if q.letters == 0 {
		return 0
	}
	return float64(q.bad) / float64(q.letters)
}

// vnLatin1 — chữ Latin-1 có trong tiếng Việt (à á â ã è é ê ì í ò ó ô õ ù ú ý…).
var vnLatin1 = func() map[rune]bool {
	m := map[rune]bool{}
	for _, r := range "àáâãèéêìíòóôõùúýÀÁÂÃÈÉÊÌÍÒÓÔÕÙÚÝ" {
		m[r] = true
	}
	return m
}()

// badRune — ký tự cho thấy phông không có bảng mã đúng: ký tự thay thế, vùng
// riêng (PUA), ký tự điều khiển, và ký tự Latin-1 mà chữ Việt Unicode không
// dùng nhưng phông TCVN3 / VNI dùng để vẽ chữ có dấu (ví dụ "ViÖt", "Vieät").
func badRune(r rune) bool {
	switch {
	case r == utf8.RuneError || r == 0xFFFE || r == 0xFFFF:
		return true
	case r >= 0xE000 && r <= 0xF8FF:
		return true
	case r < 0x20 && r != '\t' && r != '\n' && r != '\r':
		return true
	case r >= 0xA1 && r <= 0xBF:
		// ° « » © và gạch nối mềm vẫn gặp trong văn bản thường. (TCVN3 dùng « cho ô,
		// © cho â, gạch nối mềm cho ư, nhưng luôn kèm các ký tự khác trong dải này.)
		return r != 0xB0 && r != 0xAB && r != 0xBB && r != 0xA9 && r != 0xAD
	case r >= 0xC0 && r <= 0xFF:
		return !vnLatin1[r] && r != 0xD7 && r != 0xF7 // × ÷
	}
	return false
}

func pdfQuality(doc *pdfDoc) pdfQualityStats {
	var q pdfQualityStats
	for i, p := range doc.Pages {
		letters, bad := 0, 0
		for _, l := range p.Lines {
			for _, r := range l.Text {
				if unicode.IsLetter(r) {
					letters++
				}
				if badRune(r) {
					bad++
				}
			}
		}
		q.letters += letters
		q.bad += bad
		// Trang không có chữ: gần như không có chữ cái, hoặc ít chữ mà có hình (trang
		// scan thường có vài ký tự rác). Trang chỉ có một dòng tên chương vẫn là trang chữ.
		if letters < 5 || (letters < pdfPageLetters && p.Images > 0) {
			q.emptyPages = append(q.emptyPages, i+1)
		} else if float64(bad) >= 0.05*float64(letters) {
			q.badPages = append(q.badPages, i+1)
		}
	}
	return q
}

// pageList — " (trang 3, 7, 9…)" để người dùng biết chỗ cần xem; tối đa 6 trang.
func pageList(pages []int) string {
	if len(pages) == 0 {
		return ""
	}
	var parts []string
	for i, p := range pages {
		if i == 6 {
			parts = append(parts, "…")
			break
		}
		parts = append(parts, fmt.Sprint(p))
	}
	return " (trang " + strings.Join(parts, ", ") + ")"
}

// ── đầu trang, chân trang ────────────────────────────────────────────────

var (
	pageNumRe  = regexp.MustCompile(`^[\s\-–—|•·.]*(?i:trang|page|tr\.)?\s*([0-9]{1,4}|[ivxlcdm]{1,7})(\s*(/|trên|of)\s*[0-9]{1,4})?[\s\-–—|•·.]*$`)
	digitRunRe = regexp.MustCompile(`[0-9]+`)
)

// stripRunningLines bỏ dòng đầu trang / chân trang. Ứng viên: 2 dòng trên cùng
// hoặc 2 dòng dưới cùng trang (theo vị trí), nằm trong 12% mép trên / dưới, tách hẳn khỏi thân bài
// (khoảng cách tới dòng kề lớn hơn 1,5 lần khoảng cách dòng thường). Bỏ ứng
// viên lặp lại (bỏ qua chữ số) trên nhiều trang, hoặc chỉ là số trang. Trả số
// dòng đã bỏ.
func stripRunningLines(doc *pdfDoc) int {
	textPages := 0
	cand := make([]map[int]bool, len(doc.Pages)) // trang → chỉ số dòng ứng viên
	for pi, p := range doc.Pages {
		if len(p.Lines) == 0 || p.H <= 0 {
			continue
		}
		textPages++
		// Xét theo vị trí trên trang (PDFium có thể trả dòng đầu trang ở cuối danh sách).
		order := make([]int, len(p.Lines))
		for i := range order {
			order[i] = i
		}
		sort.SliceStable(order, func(a, b int) bool { return p.Lines[order[a]].Top < p.Lines[order[b]].Top })
		var gaps []float64
		for k := 1; k < len(order); k++ {
			prev, cur := p.Lines[order[k-1]], p.Lines[order[k]]
			if d := cur.Top - prev.Top; d > 0 && d < 3*maxf(prev.Size, 1) {
				gaps = append(gaps, d)
			}
		}
		spacing := median(gaps)
		apart := func(d float64) bool { return spacing == 0 || d > 1.5*spacing }
		var sizes []float64
		for _, l := range p.Lines {
			sizes = append(sizes, l.Size)
		}
		// Đầu trang, chân trang in chữ nhỏ; tên chương ở đầu trang thì chữ to: không bỏ.
		small := func(l pdfLine) bool { return l.Size <= 1.1*median(sizes) }
		cand[pi] = map[int]bool{}
		n := len(order)
		for k := 0; k < min(2, n); k++ {
			l := p.Lines[order[k]]
			if small(l) && l.Bottom <= 0.12*p.H && (k == n-1 || apart(p.Lines[order[k+1]].Top-l.Top)) {
				cand[pi][order[k]] = true
			}
		}
		for k := max(0, n-2); k < n; k++ {
			l := p.Lines[order[k]]
			if small(l) && l.Top >= 0.88*p.H && (k == 0 || apart(l.Top-p.Lines[order[k-1]].Top)) {
				cand[pi][order[k]] = true
			}
		}
	}
	key := func(s string) string {
		return strings.ToLower(collapseSpaces(digitRunRe.ReplaceAllString(s, "#")))
	}
	seen := map[string]int{}
	for pi, c := range cand {
		keys := map[string]bool{}
		for i := range c {
			keys[key(doc.Pages[pi].Lines[i].Text)] = true
		}
		for k := range keys {
			seen[k]++
		}
	}
	need := max(3, int(math.Ceil(0.3*float64(textPages))))
	removed := 0
	for pi := range doc.Pages {
		p := &doc.Pages[pi]
		kept := p.Lines[:0]
		for i, l := range p.Lines {
			if cand[pi][i] && (seen[key(l.Text)] >= need || pageNumRe.MatchString(l.Text)) {
				removed++
				continue
			}
			kept = append(kept, l)
		}
		p.Lines = kept
	}
	return removed
}

// ── dòng → đoạn ──────────────────────────────────────────────────────────

// pdfPara — 1 đoạn đã ghép, kèm vị trí để gắn bookmark.
type pdfPara struct {
	para  docxPara
	page  int
	top   float64
	size  float64
	lines int
	bold  bool
}

var sentenceEnd = regexp.MustCompile(`[.!?…:;"”»)\]]$`)

func endsSentence(s string) bool {
	return sentenceEnd.MatchString(strings.TrimSpace(s))
}

func startsLower(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return unicode.IsLower(r)
		}
		if !unicode.IsPunct(r) && !unicode.IsSpace(r) {
			return false
		}
	}
	return false
}

// joinPDFLine nối dòng sau vào đoạn: từ bị ngắt bằng gạch nối ở cuối dòng
// ("docu-" + "mentation") ghép liền, còn lại cách một dấu cách.
func joinPDFLine(text, next string) string {
	if text == "" {
		return next
	}
	t := strings.TrimRight(text, " ")
	if strings.HasSuffix(t, "­") {
		return strings.TrimSuffix(t, "­") + next
	}
	if strings.HasSuffix(t, "-") && startsLower(next) {
		before := []rune(strings.TrimSuffix(t, "-"))
		if len(before) > 0 && unicode.IsLetter(before[len(before)-1]) {
			return string(before) + next
		}
	}
	return t + " " + next
}

func median(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	s := append([]float64(nil), vals...)
	sort.Float64s(s)
	return s[len(s)/2]
}

func percentile(vals []float64, p float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	s := append([]float64(nil), vals...)
	sort.Float64s(s)
	i := int(p * float64(len(s)-1))
	return s[i]
}

// pdfParagraphs ghép các dòng thành đoạn: đoạn mới khi khoảng cách dòng lớn,
// cỡ chữ / độ đậm đổi, dòng trước ngắn và hết câu, hoặc dòng sau thụt đầu dòng.
// Đoạn bị cắt ngang giữa hai trang (chưa hết câu, trang sau bắt đầu chữ thường)
// được nối lại.
func pdfParagraphs(doc *pdfDoc) []pdfPara {
	var out []pdfPara
	for pi, p := range doc.Pages {
		if len(p.Lines) == 0 {
			continue
		}
		var gaps, x0s, x1s []float64
		for i, l := range p.Lines {
			x0s = append(x0s, l.X0)
			x1s = append(x1s, l.X1)
			if i > 0 {
				prev := p.Lines[i-1]
				if d := l.Top - prev.Top; d > 0 && d < 3*maxf(prev.Size, 1) && math.Abs(l.Size-prev.Size) < 0.5 {
					gaps = append(gaps, d)
				}
			}
		}
		spacing := median(gaps)
		left := percentile(x0s, 0.1)
		right := percentile(x1s, 0.9)

		var cur *pdfPara
		flush := func() {
			if cur != nil && strings.TrimSpace(cur.para.text) != "" {
				out = append(out, *cur)
			}
			cur = nil
		}
		for i, l := range p.Lines {
			if i > 0 && cur != nil {
				prev := p.Lines[i-1]
				d := l.Top - prev.Top
				newPara := false
				switch {
				case d <= 0 || (spacing > 0 && d > spacing*1.35) || (spacing == 0 && d > 1.8*maxf(prev.Size, 1)):
					newPara = true // khoảng trắng lớn, hoặc sang cột mới
				case math.Abs(l.Size-prev.Size) > 0.15*maxf(prev.Size, 1):
					newPara = true
				case l.Bold != prev.Bold:
					newPara = true
				case endsSentence(prev.Text) && prev.X1 < right-0.12*(right-left):
					newPara = true // dòng trước ngắn, hết câu
				case endsSentence(prev.Text) && l.X0 > left+0.8*maxf(l.Size, 1):
					newPara = true // dòng sau thụt đầu dòng
				}
				if newPara {
					flush()
				}
			}
			if cur == nil {
				cur = &pdfPara{page: pi, top: l.Top, size: l.Size, bold: true}
			}
			cur.para.text = joinPDFLine(cur.para.text, l.Text)
			cur.lines++
			cur.bold = cur.bold && l.Bold
			cur.size = maxf(cur.size, l.Size)
		}
		flush()
	}

	// Nối đoạn bị cắt giữa hai trang.
	merged := out[:0]
	for _, p := range out {
		if n := len(merged); n > 0 {
			last := &merged[n-1]
			if p.page == last.page+1 && !endsSentence(last.para.text) && startsLower(p.para.text) &&
				math.Abs(p.size-last.size) < 0.5 && p.bold == last.bold {
				last.para.text = joinPDFLine(last.para.text, p.para.text)
				last.lines += p.lines
				continue
			}
		}
		merged = append(merged, p)
	}
	for i := range merged {
		p := &merged[i]
		p.para.text = collapseSpaces(p.para.text)
		p.para.textRunes = countLetters(p.para.text)
		p.para.maxSz = int(math.Round(p.size * 2)) // nửa point, như w:sz của Word
		if p.bold {
			p.para.boldRunes = p.para.textRunes
		}
	}
	return merged
}

// ── tiêu đề: bookmark hoặc cỡ chữ ────────────────────────────────────────

// usableOutline — bookmark có trang đích hợp lệ.
func usableOutline(doc *pdfDoc) []pdfOutline {
	var out []pdfOutline
	for _, o := range doc.Outline {
		if o.Title != "" && o.Page >= 0 && o.Page < len(doc.Pages) {
			out = append(out, o)
		}
	}
	return out
}

func normTitle(s string) string {
	return strings.ToLower(collapseSpaces(strings.TrimSpace(s)))
}

// applyPDFOutline gắn bookmark vào đoạn: trên trang đích, đoạn ngắn trùng chữ
// với bookmark thành tiêu đề; không thấy thì chèn tiêu đề trước đoạn đầu tiên
// của trang đó (sau vị trí bookmark trước).
func applyPDFOutline(paras []pdfPara, outline []pdfOutline) []pdfPara {
	insertAt := map[int][]docxPara{}
	cursor := 0
	for _, o := range outline {
		style := fmt.Sprintf("Heading%d", min(o.Level, 6))
		want := normTitle(o.Title)
		// Tìm từ vị trí hiện tại tới hết trang đích.
		found := -1
		first := -1
		for i := cursor; i < len(paras); i++ {
			if paras[i].page < o.Page {
				continue
			}
			if paras[i].page > o.Page {
				break
			}
			if first < 0 {
				first = i
			}
			got := normTitle(paras[i].para.text)
			if paras[i].lines <= 3 && headingLevel(paras[i].para.style) == 0 &&
				(got == want || strings.HasPrefix(got, want) && len([]rune(got)) <= len([]rune(want))+3) {
				found = i
				break
			}
		}
		switch {
		case found >= 0:
			paras[found].para.style = style
			paras[found].para.text = o.Title
			cursor = found + 1
		case first >= 0:
			insertAt[first] = append(insertAt[first], docxPara{style: style, text: o.Title, textRunes: countLetters(o.Title)})
			cursor = first
		default:
			// Trang đích không có đoạn nào (trang ảnh): chèn trước đoạn kế tiếp.
			i := cursor
			for i < len(paras) && paras[i].page < o.Page {
				i++
			}
			insertAt[i] = append(insertAt[i], docxPara{style: style, text: o.Title, textRunes: countLetters(o.Title)})
			cursor = i
		}
	}
	out := make([]pdfPara, 0, len(paras)+len(outline))
	for i := 0; i <= len(paras); i++ {
		for _, h := range insertAt[i] {
			out = append(out, pdfPara{para: h, lines: 1})
		}
		if i < len(paras) {
			out = append(out, paras[i])
		}
	}
	return out
}

// markPDFHeadings đoán tiêu đề khi PDF không có bookmark: đoạn ngắn (≤ 3 dòng,
// ≤ 150 ký tự, không kết thúc bằng dấu chấm, phẩy) có cỡ chữ từ 1,18 lần chữ
// thân bài. Cỡ lớn nhất → Heading1, kế tiếp → Heading2, còn lại → Heading3.
// Trả số tiêu đề đã đánh dấu.
func markPDFHeadings(paras []pdfPara) int {
	weight := map[float64]int{}
	for _, p := range paras {
		weight[math.Round(p.size*2)/2] += p.para.textRunes
	}
	body, bodyW := 0.0, 0
	for sz, w := range weight {
		if w > bodyW || (w == bodyW && sz < body) {
			body, bodyW = sz, w
		}
	}
	if body == 0 {
		return 0
	}
	isCand := func(p pdfPara) bool {
		t := strings.TrimSpace(p.para.text)
		return p.size >= body*1.18 && p.lines <= 3 && len([]rune(t)) <= 150 && p.para.textRunes >= 2 &&
			!strings.ContainsAny(lastRune(t), ".,;")
	}
	sizes := map[float64]int{}
	for _, p := range paras {
		if isCand(p) {
			sizes[math.Round(p.size*2)/2]++
		}
	}
	var levels []float64
	for sz := range sizes {
		levels = append(levels, sz)
	}
	sort.Sort(sort.Reverse(sort.Float64Slice(levels)))
	// Cỡ lớn nhất chỉ có 1 đoạn, ở 2 trang đầu: đó là tên sách trên trang bìa,
	// không phải chương (nếu không, cả cuốn thành mục của một chương).
	if len(levels) > 1 && sizes[levels[0]] == 1 {
		for i := range paras {
			if isCand(paras[i]) && math.Round(paras[i].size*2)/2 == levels[0] && paras[i].page <= 1 {
				paras[i].para.style = "Title"
				levels = levels[1:]
				break
			}
		}
	}
	n := 0
	for i := range paras {
		if !isCand(paras[i]) || isTitleStyle(paras[i].para.style) {
			continue
		}
		sz := math.Round(paras[i].size*2) / 2
		lvl := 3
		for li, s := range levels {
			if s == sz {
				lvl = min(li+1, 3)
				break
			}
		}
		paras[i].para.style = fmt.Sprintf("Heading%d", lvl)
		n++
	}
	return n
}

// pdfTitle — tên sách từ metadata, bỏ giá trị rác do phần mềm xuất PDF tự ghi.
func pdfTitle(meta string) string {
	t := collapseSpaces(strings.TrimSpace(meta))
	low := strings.ToLower(t)
	switch {
	case len([]rune(t)) < 3, strings.HasPrefix(low, "microsoft word"), strings.HasPrefix(low, "untitled"),
		low == "document", low == "tài liệu", strings.HasSuffix(low, ".doc"), strings.HasSuffix(low, ".docx"),
		strings.HasSuffix(low, ".pdf"), strings.HasSuffix(low, ".indd"), strings.HasSuffix(low, ".tex"):
		return ""
	}
	return t
}

// multiColumnPages — trang (1-based) có từ 30% dòng bắt đầu ở nửa phải trang
// trong khi vẫn có dòng bắt đầu ở lề trái: dàn hai cột trở lên.
func multiColumnPages(doc *pdfDoc) []int {
	var pages []int
	for i, p := range doc.Pages {
		if p.W <= 0 || len(p.Lines) < 10 {
			continue
		}
		var x0s []float64
		for _, l := range p.Lines {
			x0s = append(x0s, l.X0)
		}
		left := percentile(x0s, 0.05)
		right := 0
		for _, l := range p.Lines {
			if l.X0 > left+0.35*p.W && l.X1-l.X0 > 0.2*p.W {
				right++
			}
		}
		if float64(right) >= 0.3*float64(len(p.Lines)) {
			pages = append(pages, i+1)
		}
	}
	return pages
}
