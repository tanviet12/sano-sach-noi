// Command sano-docx2tts là pipeline tạo sách nói tự động: đưa 1 file
// .docx của cẩm nang → trích mục lục nhiều cấp (cấp heading nhỏ nhất = chương,
// các cấp lớn hơn = tiểu mục) → chuẩn hóa cách đọc (giữ lời gốc) → sinh mô tả ảnh slide →
// pre-render TTS bằng VieNeu-TTS → đóng gói thành thư mục audiobook + metadata.json
// → (tùy chọn) gói zip chuẩn để sao lưu / chuyển máy.
//
// Ví dụ:
//
//	sano-docx2tts --input cam-nang.docx --output-dir /tmp/cam-nang/ \
//	  --voice "Thiện Minh" --output-zip /tmp/book-cam-nang.zip
//
// Xuất thêm một file M4B (mục lục chương, bìa) để nghe trên điện thoại / xe:
//
//	sano-docx2tts --input cam-nang.docx --output-dir /tmp/cam-nang/ --m4b ~/Downloads/cam-nang.m4b
//	sano-docx2tts --m4b-from-dir /tmp/cam-nang/ --m4b ~/Downloads/cam-nang.m4b
//
// Smoke / CI không có model: --tts-mode stub (sinh MP3 im lặng ngắn hợp lệ).
//
// Lệnh này chỉ đọc cờ rồi gọi package sano/internal/bookmaker (dùng chung với
// phần mềm desktop).
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sano/internal/bookmaker"
	"sano/internal/m4b"
)

func main() {
	var (
		input      = flag.String("input", "", "file .docx nguồn (bắt buộc)")
		outputDir  = flag.String("output-dir", "", "thư mục đầu ra (bắt buộc)")
		title      = flag.String("title", "", "ghi đè tiêu đề sách (mặc định: lấy từ docx)")
		author     = flag.String("author", "", "tác giả")
		desc       = flag.String("description", "", "mô tả sách")
		introText  = flag.String("intro-text", "", "đoạn intro/branding đọc đầu sách (mục đầu tiên, tên = tiêu đề sách); trống = không có intro")
		category   = flag.String("category", "", "tên hoặc slug danh mục (vd \"Kỹ năng\" → ky-nang; trống = không gán)")
		visibility = flag.String("visibility", "private", "public|private")
		tagsCSV    = flag.String("tags", "", "danh sách tag, phân tách dấu phẩy")
		voice      = flag.String("voice", bookmaker.DefaultVoice, "tên giọng preset VieNeu-TTS v3 Turbo (vd \"Thiện Minh\", \"Thái Sơn\"; xem audio_gen_batch.py --list-voices)")
		ttsMode    = flag.String("tts-mode", bookmaker.TTSModeVieNeu, "vieneu|stub (stub = MP3 im lặng cho smoke)")
		ttsOnly    = flag.String("tts-only", "", "chỉ render giọng thật cho các stem này (CSV, vd \"ch03-sec01\"); các tiểu mục còn lại sinh stub im lặng. Chỉ áp dụng khi --tts-mode vieneu")
		ttsPreview = flag.Int("tts-preview-chars", 0, "giới hạn số ký tự render giọng thật cho mỗi tiểu mục trong --tts-only (0 = đọc full; vd 1800 ≈ 1 trang để nghe thử trước khi render full)")
		dropStems  = flag.String("drop-stems", "", "loại bỏ các tiểu mục theo stem GỐC (CSV, vd \"ch01-sec01,ch02-sec01\") khỏi mục lục + zip; chương rỗng sau khi loại sẽ bị bỏ, các chương còn lại đánh số lại từ 1")
		headingNum = flag.String("heading-numbers", bookmaker.HeadingNumbersDrop, "số đầu tiêu đề khi đọc (\"1.2.3. Tên\"): drop = bỏ, chỉ đọc \"Tên\"; keep = đọc \"một chấm hai chấm ba\"")
		dropTOC    = flag.Bool("drop-toc", true, "tự bỏ trang mục lục khỏi bản đọc (tiêu đề Mục lục/Nội dung/Table of Contents, hoặc phần lớn dòng kết thúc bằng số trang); --drop-toc=false để giữ")
		cover      = flag.String("cover", "", "ảnh bìa (jpg/jpeg/png/webp); trống = tự vẽ bìa theo tên sách")
		coverFirst = flag.Bool("cover-first-image", false, "khi không có --cover: dùng ảnh đầu tiên trong file Word làm bìa thay vì tự vẽ")
		ttsPython  = flag.String("tts-python", "", "python venv VieNeu-TTS v3 (mặc định ~/VieNeu-TTS-v3/.venv/bin/python; Windows: .venv\\Scripts\\python.exe)")
		ttsScript  = flag.String("tts-script", "", "audio_gen_batch.py (mặc định: bin/../scripts/tts/ cạnh file chạy, không có thì dùng bản nhúng trong chương trình; KHÔNG tìm theo thư mục hiện tại — muốn dùng script trong repo đang sửa thì truyền cờ này)")
		ffmpegBin  = flag.String("ffmpeg", "ffmpeg", "binary ffmpeg")
		bitrate    = flag.String("bitrate", "128k", "bitrate MP3 CBR")
		stubSec    = flag.Int("stub-sec", 2, "độ dài MP3 stub (giây), chỉ dùng --tts-mode stub")
		pronFile   = flag.String("pronunciations", "", "file TSV bổ sung/ghi đè từ điển cách đọc viết tắt (mỗi dòng \"<viết tắt><TAB><cách đọc>\"; cách đọc trống = tắt mục mặc định)")
		keepTxt    = flag.Bool("keep-txt", false, "giữ lại file <stem>.txt đã nạp cho TTS trong output-dir (soát lời đọc 1:1 với audio nếu có chỗ phát sai)")
		outputZip  = flag.String("output-zip", "", "đóng gói thêm file zip chuẩn ở đường dẫn này (sao lưu / chuyển máy; xem docs/book-zip-format.md)")
		repackDir  = flag.String("repack-dir", "", "đóng gói LẠI zip từ thư mục đã render sẵn (mp3 + metadata.json) — KHÔNG cần --input docx, KHÔNG render TTS; dùng khi sửa lẻ vài tiểu mục rồi đóng gói lại. Cần --output-zip (+ --voice cho voice_id)")
		m4bOut     = flag.String("m4b", "", "xuất thêm một file .m4b (AAC, mục lục chương, bìa nhúng — nghe trên Apple Books, app sách nói Android, màn hình xe) ở đường dẫn này sau khi render xong")
		m4bFromDir = flag.String("m4b-from-dir", "", "chỉ xuất M4B từ thư mục sách đã render sẵn (metadata.json + chNN-secNN.mp3), KHÔNG cần --input, KHÔNG render; lưu ở --m4b (mặc định <thư mục>/<Tên sách>.m4b)")
		m4bBitrate = flag.String("m4b-bitrate", m4b.DefaultBitrate, "bitrate AAC mono của file M4B (64k đủ cho giọng đọc, ~29 MB/giờ)")
		verbose    = flag.Bool("verbose", false, "log chi tiết")
		genSample  = flag.String("gen-sample-docx", "", "tiện ích: ghi docx mẫu ra đường dẫn này rồi thoát")
	)
	flag.Parse()

	// Tiện ích sinh fixture docx mẫu (cho smoke/test) rồi thoát.
	if *genSample != "" {
		f, err := os.Create(*genSample)
		if err != nil {
			log.Fatalf("tạo %q: %v", *genSample, err)
		}
		defer func() { _ = f.Close() }()
		if err := bookmaker.WriteSampleDocx(f); err != nil {
			log.Fatalf("ghi docx mẫu: %v", err)
		}
		fmt.Printf("✓ Đã ghi docx mẫu: %s\n", *genSample)
		return
	}

	// Chế độ repack: đóng gói LẠI zip từ thư mục đã render (KHÔNG cần docx, KHÔNG
	// render TTS). Dùng khi sửa lẻ vài tiểu mục (render lại đoạn đọc sai) rồi đóng
	// gói lại để upload — không phải render full lại cả sách.
	if rd := strings.TrimSpace(*repackDir); rd != "" {
		if strings.TrimSpace(*outputZip) == "" {
			log.Fatalf("--repack-dir cần --output-zip <đường dẫn zip>")
		}
		logf := func(format string, a ...any) {
			if *verbose {
				fmt.Printf(format+"\n", a...)
			}
		}
		slug, err := bookmaker.RepackZip(rd, *outputZip, *voice, logf)
		if err != nil {
			log.Fatalf("repack thất bại: %v", err)
		}
		fmt.Printf("✔ Đóng gói lại zip: %s (slug=%s)\n", *outputZip, slug)
		return
	}

	// Chế độ chỉ xuất M4B từ thư mục sách đã render.
	if dir := strings.TrimSpace(*m4bFromDir); dir != "" {
		out := strings.TrimSpace(*m4bOut)
		if err := exportM4B(dir, out, *ffmpegBin, *m4bBitrate); err != nil {
			log.Fatalf("xuất M4B thất bại: %v", err)
		}
		return
	}

	if strings.TrimSpace(*input) == "" || strings.TrimSpace(*outputDir) == "" {
		fmt.Fprintln(os.Stderr, "lỗi: --input và --output-dir là bắt buộc")
		flag.Usage()
		os.Exit(2)
	}
	if !strings.EqualFold(filepath.Ext(*input), ".docx") {
		log.Fatalf("--input phải là file .docx: %q", *input)
	}
	// Chuyển output-dir sang tuyệt đối: renderVieNeu đặt cwd của python = output-dir
	// nhưng truyền path .txt theo output-dir; nếu output-dir tương đối thì path .txt
	// bị resolve sai (lồng 2 lần) → python không thấy file → không sinh WAV.
	if abs, err := filepath.Abs(*outputDir); err == nil {
		*outputDir = abs
	}
	switch *visibility {
	case "public", "private":
	default:
		log.Fatalf("--visibility không hợp lệ %q (public|private)", *visibility)
	}

	logf := func(format string, a ...any) {
		if *verbose {
			fmt.Printf(format+"\n", a...)
		}
	}

	keepNums, headingErr := bookmaker.ParseHeadingNumbers(*headingNum)
	norm, err := bookmaker.NewNormalizer(*pronFile, keepNums)
	if err != nil {
		log.Fatalf("từ điển cách đọc: %v", err)
	}
	if headingErr != nil {
		log.Fatal(headingErr)
	}

	pyDefault, scrDefault := bookmaker.TTSDefaults()
	tts := bookmaker.TTSConfig{
		Mode:         *ttsMode,
		Python:       bookmaker.FirstNonEmpty(*ttsPython, pyDefault),
		Script:       bookmaker.FirstNonEmpty(*ttsScript, scrDefault),
		Voice:        *voice,
		FFmpeg:       *ffmpegBin,
		Bitrate:      *bitrate,
		StubSec:      *stubSec,
		OnlyStems:    bookmaker.ParseStemSet(*ttsOnly),
		PreviewChars: *ttsPreview,
		KeepTxt:      *keepTxt,
		Logf:         logf,
	}
	tts.ScriptDir = filepath.Dir(tts.Script)

	opts := bookmaker.Options{
		InputDocx:       *input,
		OutputDir:       *outputDir,
		Title:           *title,
		Author:          *author,
		Description:     *desc,
		IntroText:       strings.TrimSpace(*introText),
		Category:        *category,
		Visibility:      *visibility,
		Tags:            bookmaker.SplitTags(*tagsCSV),
		TTS:             tts,
		Norm:            norm,
		DropStems:       bookmaker.ParseStemSet(*dropStems),
		DropTOC:         *dropTOC,
		CoverPath:       strings.TrimSpace(*cover),
		CoverFirstImage: *coverFirst,
		OutputZip:       strings.TrimSpace(*outputZip),
		Logf:            logf,
	}

	n, err := bookmaker.Run(opts)
	if err != nil {
		log.Fatalf("pipeline thất bại: %v", err)
	}
	fmt.Printf("\n✔ Hoàn tất: %d tiểu mục → %s\n", n, *outputDir)
	if out := strings.TrimSpace(*m4bOut); out != "" {
		if err := exportM4B(*outputDir, out, *ffmpegBin, *m4bBitrate); err != nil {
			log.Fatalf("xuất M4B thất bại: %v", err)
		}
	}
	if opts.OutputZip != "" {
		fmt.Printf("  Đã đóng gói zip chuẩn: %s\n", opts.OutputZip)
	}
}

// exportM4B xuất thư mục sách dir thành một file .m4b ở out ("" = <dir>/<Tên sách>.m4b),
// in tiến độ theo từng mười phần trăm.
func exportM4B(dir, out, ffmpeg, bitrate string) error {
	b, err := m4b.FromDir(dir)
	if err != nil {
		return err
	}
	if out == "" {
		out = filepath.Join(dir, m4b.FileName(b.Title))
	}
	fmt.Printf("Xuất M4B: %d tiểu mục → %s\n", len(b.Tracks), out)
	lastPct := -1
	res, err := m4b.Export(context.Background(), b, out, m4b.Options{
		FFmpeg:  ffmpeg,
		Bitrate: bitrate,
		Progress: func(p m4b.Progress) {
			if p.Percent/10 != lastPct/10 {
				lastPct = p.Percent
				fmt.Printf("  %3d%%  tiểu mục %d/%d\n", p.Percent, p.Track, p.Tracks)
			}
		},
	})
	if err != nil {
		return err
	}
	fmt.Printf("✔ M4B: %s (%d mốc chương, %s, %.1f MB)\n", res.Path, res.Chapters,
		(time.Duration(res.DurationSec) * time.Second).String(), float64(res.Size)/(1<<20))
	return nil
}
