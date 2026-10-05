package bookmaker

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/klippa-app/go-pdfium"
	"github.com/klippa-app/go-pdfium/enums"
	pdfiumerr "github.com/klippa-app/go-pdfium/errors"
	"github.com/klippa-app/go-pdfium/references"
	"github.com/klippa-app/go-pdfium/requests"
	"github.com/klippa-app/go-pdfium/responses"
	"github.com/klippa-app/go-pdfium/webassembly"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/experimental"
)

// Đọc PDF bằng PDFium (bộ đọc PDF của Chrome) biên dịch sang WebAssembly, chạy
// trong wazero: Go thuần, không CGO, không cần Python. Module WebAssembly không
// được gắn thư mục nào (không đọc/ghi file trên máy): Sano đưa nội dung file vào.

// pdfLine — 1 dòng chữ trên trang. Toạ độ tính từ mép TRÊN trang (point).
type pdfLine struct {
	Text        string
	X0, X1      float64
	Top, Bottom float64
	Size        float64 // cỡ chữ (point)
	Bold        bool
}

// pdfPage — chữ của 1 trang theo thứ tự đọc PDFium trả về.
type pdfPage struct {
	W, H   float64
	Lines  []pdfLine
	Images int // số hình (đối tượng ảnh) trên trang
}

// pdfOutline — 1 mục bookmark: tiêu đề, cấp (1 = ngoài cùng), trang (0-based; -1 = không rõ).
type pdfOutline struct {
	Title string
	Level int
	Page  int
}

// pdfDoc — nội dung thô trích từ PDF, chưa dựng chương/đoạn.
type pdfDoc struct {
	Title   string // metadata Title
	Pages   []pdfPage
	Outline []pdfOutline
}

// Giới hạn khi đọc PDF.
var (
	maxPDFBytes int64 = 256 << 20 // dung lượng file (nạp nguyên vào bộ nhớ)
	maxPDFPages       = 5000
	maxPDFLines       = 400 // dòng tối đa mỗi trang (trang dày đặc bất thường: cắt bớt)
	maxPDFObjs        = 400 // đối tượng tối đa duyệt mỗi trang khi đếm hình
	pdfInitWait       = 60 * time.Second
)

// ErrPDFTooLarge — PDF vượt giới hạn dung lượng / số trang.
var ErrPDFTooLarge = errors.New("file PDF quá lớn")

var pdfEngine struct {
	mu   sync.Mutex // PDFium không an toàn đa luồng: mỗi lúc chỉ 1 file
	pool pdfium.Pool
	inst pdfium.Pdfium
	err  error
	once sync.Once
}

// pdfiumInstance khởi tạo PDFium một lần cho cả tiến trình. Bản dịch máy của
// module WebAssembly được lưu trong thư mục cache người dùng để lần mở sau nhanh.
func pdfiumInstance() (pdfium.Pdfium, error) {
	pdfEngine.once.Do(func() {
		// Module PDFium cần tính năng exception-handling (như cấu hình mặc định của go-pdfium).
		rc := wazero.NewRuntimeConfig().WithCoreFeatures(api.CoreFeaturesV2 | experimental.CoreFeaturesExceptionHandling)
		if dir, err := os.UserCacheDir(); err == nil {
			if cache, err := wazero.NewCompilationCacheWithDir(filepath.Join(dir, "Sano", "pdfium")); err == nil {
				rc = rc.WithCompilationCache(cache)
			}
		}
		pool, err := webassembly.Init(webassembly.Config{
			MinIdle: 1, MaxIdle: 1, MaxTotal: 1,
			FSConfig:      wazero.NewFSConfig(), // không gắn thư mục nào
			RuntimeConfig: rc,
		})
		if err != nil {
			pdfEngine.err = fmt.Errorf("khởi động bộ đọc PDF: %w", err)
			return
		}
		inst, err := pool.GetInstance(pdfInitWait)
		if err != nil {
			pdfEngine.err = fmt.Errorf("khởi động bộ đọc PDF: %w", err)
			return
		}
		pdfEngine.pool, pdfEngine.inst = pool, inst
	})
	return pdfEngine.inst, pdfEngine.err
}

// pdfCache — kết quả đọc gần nhất (nạp mục lục, nghe thử, render đọc lại cùng file).
var pdfCache struct {
	mu    sync.Mutex
	keys  []string
	items map[string]*pdfDoc
}

const pdfCacheSize = 3

func pdfCacheKey(path string) (string, int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", 0, fmt.Errorf("mở file %q: %w", path, err)
	}
	abs, _ := filepath.Abs(path)
	return fmt.Sprintf("%s|%d|%d", abs, info.Size(), info.ModTime().UnixNano()), info.Size(), nil
}

// readPDF trích chữ, bookmark của file PDF (có cache theo đường dẫn + kích thước + giờ sửa).
func readPDF(path string) (*pdfDoc, error) {
	key, size, err := pdfCacheKey(path)
	if err != nil {
		return nil, err
	}
	if size > maxPDFBytes {
		return nil, fmt.Errorf("%w (%d MB, tối đa %d MB)", ErrPDFTooLarge, size>>20, maxPDFBytes>>20)
	}
	pdfCache.mu.Lock()
	if d, ok := pdfCache.items[key]; ok {
		pdfCache.mu.Unlock()
		return d, nil
	}
	pdfCache.mu.Unlock()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("đọc file %q: %w", path, err)
	}
	doc, err := extractPDF(data)
	if err != nil {
		return nil, err
	}

	pdfCache.mu.Lock()
	defer pdfCache.mu.Unlock()
	if pdfCache.items == nil {
		pdfCache.items = map[string]*pdfDoc{}
	}
	pdfCache.items[key] = doc
	pdfCache.keys = append(pdfCache.keys, key)
	if len(pdfCache.keys) > pdfCacheSize {
		delete(pdfCache.items, pdfCache.keys[0])
		pdfCache.keys = pdfCache.keys[1:]
	}
	return doc, nil
}

// resetPDFium bỏ phiên PDFium hiện tại (sau khi một file hỏng làm PDFium lỗi
// giữa chừng) và lấy phiên mới, để file sau vẫn đọc được mà không phải mở lại app.
// Gọi khi đang giữ pdfEngine.mu.
func resetPDFium() {
	if pdfEngine.inst != nil {
		_ = pdfEngine.inst.Kill()
	}
	pdfEngine.inst = nil
	if pdfEngine.pool == nil {
		return
	}
	if inst, err := pdfEngine.pool.GetInstance(pdfInitWait); err == nil {
		pdfEngine.inst = inst
	} else {
		pdfEngine.err = fmt.Errorf("khởi động lại bộ đọc PDF: %w", err)
	}
}

// extractPDF đọc nội dung PDF (bytes) bằng PDFium.
func extractPDF(data []byte) (doc *pdfDoc, err error) {
	if _, err := pdfiumInstance(); err != nil {
		return nil, err
	}
	pdfEngine.mu.Lock()
	defer pdfEngine.mu.Unlock()
	inst := pdfEngine.inst
	if inst == nil {
		if pdfEngine.err != nil {
			return nil, pdfEngine.err
		}
		return nil, errors.New("bộ đọc PDF chưa sẵn sàng")
	}
	defer func() {
		if r := recover(); r != nil {
			resetPDFium()
			doc, err = nil, fmt.Errorf("file PDF hỏng, bộ đọc PDF không xử lý được: %v", r)
		}
	}()
	doc, err = extractWith(inst, data)
	if err != nil && !errors.Is(err, ErrProtectedFile) && !errors.Is(err, ErrPDFTooLarge) && !errors.Is(err, errPDFFormat) {
		resetPDFium() // lỗi lạ: có thể PDFium đã hỏng trạng thái
	}
	return doc, err
}

// errPDFFormat — file không đúng định dạng PDF hoặc hỏng.
var errPDFFormat = errors.New("file PDF hỏng hoặc không đúng định dạng")

func extractWith(inst pdfium.Pdfium, data []byte) (*pdfDoc, error) {

	opened, err := inst.OpenDocument(&requests.OpenDocument{File: &data})
	if err != nil {
		if errors.Is(err, pdfiumerr.ErrPassword) {
			return nil, ErrProtectedFile
		}
		if errors.Is(err, pdfiumerr.ErrFormat) || errors.Is(err, pdfiumerr.ErrFile) {
			return nil, errPDFFormat
		}
		if errors.Is(err, pdfiumerr.ErrSecurity) {
			return nil, ErrProtectedFile
		}
		return nil, fmt.Errorf("mở PDF: %w", err)
	}
	doc := opened.Document
	defer func() { _, _ = inst.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: doc}) }()

	// PDF có khoá chủ sở hữu cấm trích chữ (cả cho người khiếm thị): từ chối như
	// file Word có khoá. Cho phép trích chữ hỗ trợ tiếp cận (bit 10) thì vẫn đọc.
	if rev, err := inst.FPDF_GetSecurityHandlerRevision(&requests.FPDF_GetSecurityHandlerRevision{Document: doc}); err == nil && rev.SecurityHandlerRevision != -1 {
		if perm, err := inst.FPDF_GetDocPermissions(&requests.FPDF_GetDocPermissions{Document: doc}); err == nil &&
			!perm.CopyOrExtractText && !perm.ExtractTextAndGraphics {
			return nil, ErrProtectedFile
		}
	}

	pc, err := inst.FPDF_GetPageCount(&requests.FPDF_GetPageCount{Document: doc})
	if err != nil {
		return nil, fmt.Errorf("đọc số trang PDF: %w", err)
	}
	if pc.PageCount > maxPDFPages {
		return nil, fmt.Errorf("%w (%d trang, tối đa %d trang): chia thành nhiều tập", ErrPDFTooLarge, pc.PageCount, maxPDFPages)
	}

	out := &pdfDoc{Pages: make([]pdfPage, 0, pc.PageCount)}
	if mt, err := inst.FPDF_GetMetaText(&requests.FPDF_GetMetaText{Document: doc, Tag: "Title"}); err == nil {
		out.Title = strings.TrimSpace(mt.Value)
	}
	if bm, err := inst.GetBookmarks(&requests.GetBookmarks{Document: doc}); err == nil {
		var walk func([]responses.GetBookmarksBookmark, int)
		walk = func(list []responses.GetBookmarksBookmark, level int) {
			for _, b := range list {
				page := -1
				if b.DestInfo != nil {
					page = b.DestInfo.PageIndex
				}
				out.Outline = append(out.Outline, pdfOutline{Title: collapseSpaces(strings.TrimSpace(b.Title)), Level: level, Page: page})
				if level < 6 {
					walk(b.Children, level+1)
				}
			}
		}
		walk(bm.Bookmarks, 1)
	}

	for i := 0; i < pc.PageCount; i++ {
		p, err := readPDFPage(inst, doc, i)
		if err != nil {
			p = pdfPage{} // trang hỏng: coi như trang không có chữ
		}
		out.Pages = append(out.Pages, p)
	}
	return out, nil
}

func readPDFPage(inst pdfium.Pdfium, doc references.FPDF_DOCUMENT, index int) (pdfPage, error) {
	var pg pdfPage
	lp, err := inst.FPDF_LoadPage(&requests.FPDF_LoadPage{Document: doc, Index: index})
	if err != nil {
		return pg, err
	}
	page := lp.Page
	defer func() { _, _ = inst.FPDF_ClosePage(&requests.FPDF_ClosePage{Page: page}) }()
	ref := requests.Page{ByReference: &page}

	if h, err := inst.FPDF_GetPageHeightF(&requests.FPDF_GetPageHeightF{Page: ref}); err == nil {
		pg.H = float64(h.PageHeight)
	}
	if w, err := inst.FPDF_GetPageWidthF(&requests.FPDF_GetPageWidthF{Page: ref}); err == nil {
		pg.W = float64(w.PageWidth)
	}
	if n, err := inst.FPDFPage_CountObjects(&requests.FPDFPage_CountObjects{Page: ref}); err == nil {
		for j := 0; j < min(n.Count, maxPDFObjs); j++ {
			o, err := inst.FPDFPage_GetObject(&requests.FPDFPage_GetObject{Page: ref, Index: j})
			if err != nil {
				continue
			}
			if t, err := inst.FPDFPageObj_GetType(&requests.FPDFPageObj_GetType{PageObject: o.PageObject}); err == nil && t.Type == enums.FPDF_PAGEOBJ_IMAGE {
				pg.Images++
			}
		}
	}

	tp, err := inst.FPDFText_LoadPage(&requests.FPDFText_LoadPage{Page: ref})
	if err != nil {
		return pg, err
	}
	text := tp.TextPage
	defer func() { _, _ = inst.FPDFText_ClosePage(&requests.FPDFText_ClosePage{TextPage: text}) }()
	cc, err := inst.FPDFText_CountChars(&requests.FPDFText_CountChars{TextPage: text})
	if err != nil || cc.Count <= 0 {
		return pg, err
	}
	gt, err := inst.FPDFText_GetText(&requests.FPDFText_GetText{TextPage: text, StartIndex: 0, Count: cc.Count})
	if err != nil {
		return pg, err
	}
	// Chỉ số ký tự PDFium = chỉ số rune (mỗi ký tự là 1 điểm mã Unicode).
	runes := []rune(gt.Text)
	box := func(i int) (l, r, top, bottom float64, ok bool) {
		b, err := inst.FPDFText_GetCharBox(&requests.FPDFText_GetCharBox{TextPage: text, Index: i})
		if err != nil {
			return 0, 0, 0, 0, false
		}
		return b.Left, b.Right, pg.H - b.Top, pg.H - b.Bottom, true
	}
	size := func(i int) float64 {
		s, err := inst.FPDFText_GetFontSize(&requests.FPDFText_GetFontSize{TextPage: text, Index: i})
		if err != nil {
			return 0
		}
		return s.FontSize
	}
	bold := func(i int) bool {
		if w, err := inst.FPDFText_GetFontWeight(&requests.FPDFText_GetFontWeight{TextPage: text, Index: i}); err == nil && w.FontWeight >= 600 {
			return true
		}
		if f, err := inst.FPDFText_GetFontInfo(&requests.FPDFText_GetFontInfo{TextPage: text, Index: i}); err == nil {
			n := strings.ToLower(f.FontName)
			return strings.Contains(n, "bold") || strings.Contains(n, "black") || strings.Contains(n, "heavy") || strings.Contains(n, "semibold")
		}
		return false
	}

	start := 0
	for i := 0; i <= len(runes) && i <= cc.Count; i++ {
		if i < len(runes) && runes[i] != '\n' {
			continue
		}
		seg := runes[start:i]
		first, last := -1, -1
		for k, r := range seg {
			if r != '\r' && r != ' ' && r != '\t' && r != 0xA0 {
				if first < 0 {
					first = k
				}
				last = k
			}
		}
		if first >= 0 && len(pg.Lines) < maxPDFLines {
			i0, i1 := start+first, start+last
			mid := (i0 + i1) / 2
			ln := pdfLine{Text: strings.TrimSpace(strings.ReplaceAll(string(seg), "\r", ""))}
			l0, _, t0, b0, ok0 := box(i0)
			_, r1, t1, b1, ok1 := box(i1)
			if ok0 && ok1 {
				ln.X0, ln.X1 = l0, r1
				ln.Top, ln.Bottom = minf(t0, t1), maxf(b0, b1)
			}
			sizes := []float64{size(i0), size(mid), size(i1)}
			ln.Size = median3(sizes[0], sizes[1], sizes[2])
			ln.Bold = bold(mid)
			pg.Lines = append(pg.Lines, ln)
		}
		start = i + 1
	}
	return pg, nil
}

func minf(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxf(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func median3(a, b, c float64) float64 {
	if a > b {
		a, b = b, a
	}
	if b > c {
		b = c
	}
	if a > b {
		return a
	}
	return b
}
