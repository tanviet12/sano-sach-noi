// Package cover vẽ bìa sách mặc định thành ảnh PNG khi sách chưa có ảnh bìa.
// Thiết kế khớp component `BookCover.vue` của phần mềm: màu chọn theo tên sách (cùng hàm băm
// FNV-1a, cùng thứ tự bảng màu nên giao diện và file ảnh ra cùng một màu), gáy sách bên
// trái, vầng sáng góc trên, nhãn "SÁCH NÓI", sóng âm của logo, tên sách chữ có chân,
// tác giả chữ hoa. Font Noto Serif / Noto Sans (SIL OFL 1.1, xem fonts/OFL.txt).
package cover

import (
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"math"
	"strings"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
	"golang.org/x/text/unicode/norm"
)

//go:embed fonts/NotoSerif-SemiBold.ttf
var serifTTF []byte

//go:embed fonts/NotoSans-SemiBold.ttf
var sansTTF []byte

// palette — nền, màu nhấn (nhãn, sóng âm), màu chữ phụ (tác giả).
// Giá trị = màu Tailwind tương ứng trong BookCover.vue; giữ ĐÚNG thứ tự.
type palette struct{ bg, accent, sub color.RGBA }

var palettes = []palette{
	{hex(0x1e1b4b), hex(0xa5b4fc), hex(0xc7d2fe)}, // indigo-950 / indigo-300 / indigo-200
	{hex(0x292524), hex(0xfcd34d), hex(0xe7e5e4)}, // stone-800 / amber-300 / stone-200
	{hex(0x881337), hex(0xfda4af), hex(0xffe4e6)}, // rose-900 / rose-300 / rose-100
	{hex(0x92400e), hex(0xfde68a), hex(0xfef3c7)}, // amber-800 / amber-200 / amber-100
	{hex(0x1e293b), hex(0x7dd3fc), hex(0xe2e8f0)}, // slate-800 / sky-300 / slate-200
	{hex(0x4c1d95), hex(0xc4b5fd), hex(0xede9fe)}, // violet-900 / violet-300 / violet-100
	{hex(0x115e59), hex(0x99f6e4), hex(0xccfbf1)}, // teal-800 / teal-200 / teal-100
	{hex(0x991b1b), hex(0xfecaca), hex(0xfee2e2)}, // red-800 / red-200 / red-100
}

func hex(v uint32) color.RGBA {
	return color.RGBA{R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v), A: 0xff}
}

// PaletteIndex — FNV-1a 32 bit trên từng code point, giống hệt BookCover.vue.
func PaletteIndex(title string) int {
	h := uint32(2166136261)
	for _, r := range title {
		h = (h ^ uint32(r)) * 16777619
	}
	return int(h % uint32(len(palettes)))
}

var (
	fontsOnce sync.Once
	serifFont *opentype.Font
	sansFont  *opentype.Font
	fontsErr  error
)

func loadFonts() error {
	fontsOnce.Do(func() {
		if serifFont, fontsErr = opentype.Parse(serifTTF); fontsErr != nil {
			return
		}
		sansFont, fontsErr = opentype.Parse(sansTTF)
	})
	return fontsErr
}

func face(f *opentype.Font, size float64) (font.Face, error) {
	return opentype.NewFace(f, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
}

// Render vẽ bìa w×h. Tỷ lệ gợi ý: 3:4 (vd 1200×1600) cho gói zip / thư viện,
// vuông (1400×1400) cho M4B (trình phát trên xe, điện thoại hiển thị bìa vuông).
func Render(title, author string, w, h int) (image.Image, error) {
	if w < 64 || h < 64 {
		return nil, fmt.Errorf("kích thước bìa quá nhỏ: %dx%d", w, h)
	}
	if err := loadFonts(); err != nil {
		return nil, fmt.Errorf("nạp font bìa: %w", err)
	}
	title = strings.TrimSpace(clipRunes(norm.NFC.String(title), maxTextRunes))
	author = strings.TrimSpace(clipRunes(norm.NFC.String(author), maxTextRunes))
	if title == "" {
		title = "Sách nói"
	}
	p := palettes[PaletteIndex(title)]
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: p.bg}, image.Point{}, draw.Src)

	u := float64(w) / 100 // 1 đơn vị = 1% bề rộng (khớp cqw trong BookCover.vue)

	// Vầng sáng góc trên phải: tâm (100%, 0), bán kính ~75% bề rộng, trắng 10% mờ dần.
	glow(img, float64(w)*0.9, float64(h)*0.05, float64(w)*0.75, 0.10)
	// Gáy sách: dải tối 7% + vệt sáng 1px.
	fillRect(img, 0, 0, int(7*u), h, color.RGBA{A: 64})
	fillRect(img, int(7*u), 0, int(7*u)+max(1, int(u*0.25)), h, color.RGBA{R: 255, G: 255, B: 255, A: 38})

	padL, padT, padR, padB := 16*u, 9*u, 9*u, 9*u

	// Nhãn "SÁCH NÓI" + gạch dưới.
	labelFace, err := face(sansFont, 5.5*u)
	if err != nil {
		return nil, err
	}
	defer labelFace.Close()
	labelBase := padT + 5.5*u
	drawTracked(img, labelFace, "SÁCH NÓI", padL, labelBase, 0.2*5.5*u, p.accent)
	fillRect(img, int(padL), int(labelBase+3*u), int(padL+18*u), int(labelBase+3*u)+max(1, int(u*0.3)), withAlpha(p.accent, 153))

	// Sóng âm góc trên phải (5 vạch như logo).
	barsH, barW, gap := 16*u, 2.6*u, 2.2*u
	heights := []float64{0.35, 0.70, 1.0, 0.60, 0.40}
	x := float64(w) - padR - (5*barW + 4*gap)
	cy := padT + barsH/2
	for _, hh := range heights {
		bh := barsH * hh
		roundedBar(img, x, cy-bh/2, barW, bh, withAlpha(p.accent, 178))
		x += barW + gap
	}

	// Tác giả (dưới cùng) — đo trước để biết chỗ dành cho tên sách.
	bottom := float64(h) - padB
	if author != "" {
		af, err := face(sansFont, 5.5*u)
		if err != nil {
			return nil, err
		}
		defer af.Close()
		txt := truncate(af, strings.ToUpper(author), float64(w)-padL-padR, 0.08*5.5*u)
		drawTracked(img, af, txt, padL, bottom, 0.08*5.5*u, withAlpha(p.sub, 204))
		bottom -= 5.5*u + 5*u
	}

	// Tên sách: chữ có chân, tối đa 4 dòng, tự thu nhỏ tới khi vừa. Căn đáy.
	maxW := float64(w) - padL - padR
	top := padT + barsH + 6*u
	size := 12.5 * u
	var lines []string
	var tf font.Face
	for ; size >= 6*u; size -= 0.5 * u {
		if tf != nil {
			tf.Close()
		}
		if tf, err = face(serifFont, size); err != nil {
			return nil, err
		}
		lines = wrap(tf, title, maxW)
		if len(lines) <= 4 && float64(len(lines))*size*1.15 <= bottom-top {
			break
		}
	}
	defer tf.Close()
	if len(lines) > 4 {
		lines = lines[:4]
		lines[3] = truncate(tf, lines[3]+"…", maxW, 0)
	}
	lh := size * 1.15
	y := bottom - float64(len(lines)-1)*lh
	for _, ln := range lines {
		drawTracked(img, tf, ln, padL, y, 0, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		y += lh
	}
	return img, nil
}

// maxTextRunes — tên sách / tác giả dài hơn thì cắt trước khi vẽ (bìa chỉ hiện
// tối đa 4 dòng; tên cực dài từ file Word không được làm việc ngắt dòng chậm).
const maxTextRunes = 300

// clipRunes cắt s còn tối đa n ký tự (rune).
func clipRunes(s string, n int) string {
	i := 0
	for pos := range s {
		if i == n {
			return s[:pos]
		}
		i++
	}
	return s
}

// WritePNG vẽ bìa rồi mã hoá PNG ra w.
func WritePNG(out io.Writer, title, author string, w, h int) error {
	img, err := Render(title, author, w, h)
	if err != nil {
		return err
	}
	return png.Encode(out, img)
}

// ---- vẽ phụ trợ ----

func withAlpha(c color.RGBA, a uint8) color.RGBA { return color.RGBA{c.R, c.G, c.B, a} }

// blend trộn màu c (alpha thẳng) lên điểm (x, y).
func blend(img *image.RGBA, x, y int, c color.RGBA, cover float64) {
	if !(image.Point{x, y}.In(img.Rect)) || cover <= 0 {
		return
	}
	a := float64(c.A) / 255 * math.Min(cover, 1)
	i := img.PixOffset(x, y)
	pix := img.Pix[i : i+3 : i+3]
	pix[0] = uint8(float64(pix[0])*(1-a) + float64(c.R)*a)
	pix[1] = uint8(float64(pix[1])*(1-a) + float64(c.G)*a)
	pix[2] = uint8(float64(pix[2])*(1-a) + float64(c.B)*a)
}

func fillRect(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			blend(img, x, y, c, 1)
		}
	}
}

// glow — vòng tròn trắng mờ dần ra ngoài (thay cho blur-2xl trong BookCover.vue).
func glow(img *image.RGBA, cx, cy, r, strength float64) {
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			d := math.Hypot(float64(x)-cx, float64(y)-cy) / r
			if d < 1 {
				t := 1 - d
				blend(img, x, y, color.RGBA{255, 255, 255, 255}, strength*t*t)
			}
		}
	}
}

// roundedBar — hình chữ nhật bo tròn hai đầu (vạch sóng âm), khử răng cưa đơn giản.
func roundedBar(img *image.RGBA, x, y, w, h float64, c color.RGBA) {
	r := w / 2
	for py := int(y) - 1; py <= int(y+h)+1; py++ {
		for px := int(x) - 1; px <= int(x+w)+1; px++ {
			fx, fy := float64(px)+0.5, float64(py)+0.5
			cx := math.Min(math.Max(fx, x+r), x+w-r)
			cy := math.Min(math.Max(fy, y+r), y+h-r)
			d := math.Hypot(fx-cx, fy-cy)
			blend(img, px, py, c, r-d+0.5)
		}
	}
}

func advance(f font.Face, s string, tracking float64) float64 {
	n := 0
	for range s {
		n++
	}
	return float64(font.MeasureString(f, s))/64 + tracking*float64(max(n-1, 0))
}

// drawTracked vẽ chuỗi tại (x, baseline) với giãn chữ `tracking` px.
func drawTracked(img *image.RGBA, f font.Face, s string, x, baseline, tracking float64, c color.RGBA) {
	// color.RGBA là màu đã nhân alpha — chữ bán trong suốt phải dùng NRGBA (alpha thẳng).
	d := &font.Drawer{Dst: img, Src: image.NewUniform(color.NRGBA{c.R, c.G, c.B, c.A}), Face: f}
	d.Dot = fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(baseline * 64)}
	if tracking == 0 {
		d.DrawString(s)
		return
	}
	for _, r := range s {
		d.DrawString(string(r))
		d.Dot.X += fixed.Int26_6(tracking * 64)
	}
}

// wrap ngắt dòng theo từ cho vừa maxW. Từ dài hơn cả dòng thì để nguyên một dòng.
func wrap(f font.Face, s string, maxW float64) []string {
	var lines []string
	cur := ""
	for _, word := range strings.Fields(s) {
		try := word
		if cur != "" {
			try = cur + " " + word
		}
		if cur != "" && advance(f, try, 0) > maxW {
			lines = append(lines, cur)
			cur = word
			continue
		}
		cur = try
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	return lines
}

// truncate cắt chuỗi thêm "…" cho vừa maxW. Một lượt cộng dồn độ rộng từng
// glyph (không đo lại cả chuỗi mỗi lần bớt một ký tự → không chạy bình phương).
func truncate(f font.Face, s string, maxW, tracking float64) string {
	if advance(f, s, tracking) <= maxW {
		return s
	}
	rs := []rune(strings.TrimSuffix(s, "…"))
	ell, _ := f.GlyphAdvance('…')
	budget := maxW - float64(ell)/64 - tracking
	w := 0.0
	n := 0
	for i, r := range rs {
		adv, _ := f.GlyphAdvance(r)
		gw := float64(adv) / 64
		if i > 0 {
			gw += tracking
		}
		if w+gw > budget {
			break
		}
		w += gw
		n = i + 1
	}
	// Kerning có thể làm lệch vài phần px: kiểm lại bằng đo thật, lùi tối đa vài ký tự.
	for back := 0; n > 0 && back < 8; back++ {
		t := strings.TrimSpace(string(rs[:n])) + "…"
		if advance(f, t, tracking) <= maxW {
			return t
		}
		n--
	}
	if n > 0 {
		return strings.TrimSpace(string(rs[:n])) + "…"
	}
	return "…"
}
