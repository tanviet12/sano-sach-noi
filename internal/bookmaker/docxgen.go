package bookmaker

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"strings"
)

// WriteSampleDocx ghi 1 file .docx mẫu (~3 chương, có Heading 1/2, đoạn văn, 1 ảnh)
// vào w. Dùng cho fixture test + smoke pipeline — nội dung gốc tự viết (tránh
// bản quyền), có sẵn số La Mã + viết tắt để kiểm tra bước chuẩn hóa.
func WriteSampleDocx(w io.Writer) error {
	zw := zip.NewWriter(w)

	type entry struct {
		name string
		data []byte
	}
	entries := []entry{
		{"[Content_Types].xml", []byte(sampleContentTypes)},
		{"_rels/.rels", []byte(sampleRootRels)},
		{"word/document.xml", []byte(sampleDocumentXML)},
		{"word/_rels/document.xml.rels", []byte(sampleDocRels)},
		{"word/media/image1.png", samplePNG()},
	}
	for _, e := range entries {
		fw, err := zw.Create(e.name)
		if err != nil {
			return fmt.Errorf("zip create %q: %w", e.name, err)
		}
		if _, err := fw.Write(e.data); err != nil {
			return fmt.Errorf("zip write %q: %w", e.name, err)
		}
	}
	return zw.Close()
}

// samplePNG sinh ảnh PNG nhỏ màu đỏ thương hiệu (#c60505) làm slide mẫu.
func samplePNG() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	red := color.RGBA{R: 0xc6, G: 0x05, B: 0x05, A: 0xff}
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img.Set(x, y, red)
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

const sampleContentTypes = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Default Extension="png" ContentType="image/png"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`

const sampleRootRels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`

const sampleDocRels = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId100" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="media/image1.png"/>
</Relationships>`

// sampleDocumentXML — thân tài liệu mẫu: 1 Title + 3 chương (Heading1), mỗi
// chương vài tiểu mục (Heading2) + đoạn văn; chương I có 1 ảnh minh họa.
var sampleDocumentXML = buildSampleDocumentXML()

func buildSampleDocumentXML() string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessing" xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture"><w:body>`)

	styled := func(style, text string) {
		b.WriteString(`<w:p><w:pPr><w:pStyle w:val="` + style + `"/></w:pPr><w:r><w:t xml:space="preserve">` + xmlEscape(text) + `</w:t></w:r></w:p>`)
	}
	body := func(text string) {
		b.WriteString(`<w:p><w:r><w:t xml:space="preserve">` + xmlEscape(text) + `</w:t></w:r></w:p>`)
	}
	imagePara := func(relID string) {
		b.WriteString(`<w:p><w:r><w:drawing><wp:inline><wp:extent cx="1000000" cy="1000000"/><a:graphic><a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/picture"><pic:pic><pic:blipFill><a:blip r:embed="` + relID + `"/></pic:blipFill></pic:pic></a:graphicData></a:graphic></wp:inline></w:drawing></w:r></w:p>`)
	}

	styled("Title", "Cẩm Nang Mô Hình Kinh Doanh Mẫu")

	styled("Heading1", "Chương I: Nền tảng tư duy")
	styled("Heading2", "Tiểu mục 1: Khởi đầu đúng hướng")
	body("Mọi mô hình kinh doanh đều bắt đầu từ một bài toán thực tế của khách hàng. Khi hiểu rõ vấn đề, vd. nhu cầu chưa được phục vụ, ta mới thiết kế giải pháp phù hợp.")
	body("Hình dưới đây minh họa khung mô hình kinh doanh chín ô.")
	imagePara("rId100")
	styled("Heading2", "Tiểu mục 2: Khách hàng là trung tâm")
	body("Đặt khách hàng vào trung tâm giúp doanh nghiệp ở TP.HCM cũng như các tỉnh khác giữ được lợi thế cạnh tranh lâu dài.")

	styled("Heading1", "Chương II: Dòng tiền và tăng trưởng")
	styled("Heading2", "Tiểu mục 1: Quản trị dòng tiền")
	body("Dòng tiền là mạch máu của doanh nghiệp. Một công ty có lãi trên sổ sách vẫn có thể sụp đổ nếu thiếu tiền mặt, vd. khi công nợ kéo dài.")
	styled("Heading2", "Tiểu mục 2: Tái đầu tư hợp lý")
	body("Lợi nhuận giữ lại nên được tái đầu tư vào những kênh tạo ra giá trị bền vững, tránh dàn trải v.v.")

	styled("Heading1", "Chương III: Đội ngũ và văn hóa")
	body("Chương này bàn về con người, yếu tố quyết định sự bền vững của mọi tổ chức.")
	styled("Heading2", "Tiểu mục 1: Tuyển đúng người")
	body("Tuyển đúng người quan trọng hơn tuyển nhanh. Một đội ngũ nhỏ nhưng phù hợp văn hóa sẽ đi xa hơn một đội ngũ đông mà rời rạc.")

	b.WriteString(`</w:body></w:document>`)
	return b.String()
}

func xmlEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
	return r.Replace(s)
}
