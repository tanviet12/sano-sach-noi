package main

// Từ điển cách đọc (wireframe D12): bộ chuẩn có sẵn, từ điển chung (Cài đặt), từ
// điển riêng của cuốn (Tạo sách → Nghe thử, Sửa sách → tab Từ điển).

import (
	"log"
	"sort"

	"sano/internal/bookmaker"
)

// DictEntry — một dòng từ điển hiện ở Cài đặt → Từ điển chung.
type DictEntry struct {
	Word    string `json:"word"`
	Reading string `json:"reading"`
	Builtin bool   `json:"builtin"` // có trong bộ chuẩn
	Mine    bool   `json:"mine"`    // người dùng tự thêm / ghi đè (xoá được)
	Default string `json:"default"` // cách đọc chuẩn khi đã ghi đè
}

// Pronunciations — bộ chuẩn + từ điển chung, gộp theo từ (từ tự thêm lên đầu).
func (a *App) Pronunciations() ([]DictEntry, error) {
	user, err := a.lib.GlobalDict()
	if err != nil {
		return nil, err
	}
	def := bookmaker.DefaultPronunciations()
	out := make([]DictEntry, 0, len(def)+len(user))
	for w, r := range user {
		out = append(out, DictEntry{Word: w, Reading: r, Builtin: def[w] != "", Mine: true, Default: def[w]})
	}
	for w, r := range def {
		if _, ok := user[w]; !ok && r != "" {
			out = append(out, DictEntry{Word: w, Reading: r, Builtin: true})
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Mine != out[j].Mine {
			return out[i].Mine
		}
		return out[i].Word < out[j].Word
	})
	return out, nil
}

// SetGlobalPronunciation thêm / sửa một từ trong từ điển chung (áp cho sách tạo sau).
func (a *App) SetGlobalPronunciation(word, reading string) error {
	return a.lib.SetGlobalWord(word, reading)
}

// DeleteGlobalPronunciation xoá từ tự thêm khỏi từ điển chung.
func (a *App) DeleteGlobalPronunciation(word string) error {
	return a.lib.DeleteGlobalWord(word)
}

// BookPronunciations — từ điển riêng của một cuốn.
func (a *App) BookPronunciations(slug string) (map[string]string, error) {
	return a.lib.BookDict(slug)
}

// SetBookPronunciation thêm / sửa một từ trong từ điển của cuốn. Các mục có từ
// đó cần đọc lại (giao diện đánh dấu); gói zip đóng lại ngay ở nền.
func (a *App) SetBookPronunciation(slug, word, reading string) error {
	if err := a.lib.SetBookWord(slug, word, reading); err != nil {
		return err
	}
	a.repackLater(slug)
	return nil
}

// DeleteBookPronunciation xoá một từ khỏi từ điển của cuốn.
func (a *App) DeleteBookPronunciation(slug, word string) error {
	if err := a.lib.DeleteBookWord(slug, word); err != nil {
		return err
	}
	a.repackLater(slug)
	return nil
}

// repackLater đóng gói lại zip ở nền (từ điển trong gói theo kịp bản trong thư mục).
func (a *App) repackLater(slug string) {
	go func() {
		if err := a.lib.Repack(slug, ""); err != nil {
			log.Printf("đóng gói lại %s sau khi sửa từ điển: %v", slug, err)
		}
	}()
}

// CountWords đếm số chỗ có từng từ trong file Word / EPUB / PDF (cho "N chỗ trong sách").
func (a *App) CountWords(path string, words []string) (map[string]int, error) {
	if _, err := describeDocx(path); err != nil {
		return nil, err
	}
	return bookmaker.CountWordsInDocx(path, words)
}

// globalDict — từ điển chung; đọc lỗi thì coi như trống (vẫn đọc bằng bộ chuẩn).
func (a *App) globalDict() map[string]string {
	m, err := a.lib.GlobalDict()
	if err != nil {
		log.Printf("từ điển chung: %v", err)
		return nil
	}
	return m
}

// normalizerFor dựng bộ chuẩn hóa: bộ chuẩn + từ điển chung + từ điển của cuốn.
func (a *App) normalizerFor(keepHeadingNumbers bool, book map[string]string) (*bookmaker.Normalizer, error) {
	return bookmaker.NewNormalizerWith(keepHeadingNumbers, a.globalDict(), book)
}
