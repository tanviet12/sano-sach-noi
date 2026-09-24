package bookmaker

import "testing"

func TestIsTOCTitle(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"Mục lục", true},
		{"MỤC LỤC", true},
		{"Mục lục:", true},
		{"Nội dung", true},
		{"Table of Contents", true},
		{"Contents", true},
		{"1. Mục lục", true},
		{"Nội dung chính của chương", false},
		{"Lời mở đầu", false},
	}
	for _, tc := range cases {
		if got := isTOCTitle(tc.in); got != tc.want {
			t.Errorf("isTOCTitle(%q) = %v, muốn %v", tc.in, got, tc.want)
		}
	}
}

func TestIsTOCBody(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{
			name: "dòng kết thúc bằng số trang",
			in:   "1. Mở đầu 3\n\n1.1. Bối cảnh 5\n\n1.2. Mục tiêu . 8\n\n2. Phương pháp 12\n\nPhụ lục ........ 40",
			want: true,
		},
		{
			name: "văn xuôi bình thường",
			in:   "Đây là đoạn văn thứ nhất.\n\nĐoạn thứ hai nói về kế hoạch năm 2025.\n\nĐoạn ba kết thúc bình thường.\n\nĐoạn bốn cũng vậy.",
			want: false,
		},
		{
			name: "quá ít dòng",
			in:   "Chương một 3\n\nChương hai 9",
			want: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isTOCBody(tc.in); got != tc.want {
				t.Errorf("isTOCBody = %v, muốn %v", got, tc.want)
			}
		})
	}
}

func TestDropTOCSections(t *testing.T) {
	book := &Book{Chapters: []Chapter{
		{Title: "Mục lục", Sections: []Section{{Title: "Mục lục", Text: "1. Mở đầu 3\n\n2. Thân bài 9"}}},
		{Title: "1. Mở đầu", Sections: []Section{
			{Title: "1. Mở đầu", Text: "Nội dung mở đầu."},
			{Title: "1.1. Danh sách", Text: "Phần A 1\n\nPhần B 2\n\nPhần C 3\n\nPhần D 4"},
		}},
	}}
	dropped := dropTOCSections(book)
	if len(dropped) != 2 {
		t.Fatalf("phải bỏ 2 tiểu mục, bỏ %d: %+v", len(dropped), dropped)
	}
	if len(book.Chapters) != 1 || len(book.Chapters[0].Sections) != 1 {
		t.Fatalf("còn lại sai: %+v", book.Chapters)
	}
	if got := book.Chapters[0].Sections[0].Title; got != "1. Mở đầu" {
		t.Errorf("tiểu mục còn lại = %q", got)
	}
	if dropped[0].Reason != "tiêu đề" || dropped[1].Reason != "dòng kết thúc bằng số trang" {
		t.Errorf("lý do sai: %+v", dropped)
	}
}

func TestDropTOCSections_NeverDropsEverything(t *testing.T) {
	book := &Book{Chapters: []Chapter{
		{Title: "Nội dung", Sections: []Section{{Title: "Nội dung", Text: "Toàn bộ sách không có heading."}}},
	}}
	if dropped := dropTOCSections(book); dropped != nil {
		t.Errorf("không được bỏ tiểu mục duy nhất: %+v", dropped)
	}
	if len(book.Chapters) != 1 {
		t.Errorf("book bị đổi: %+v", book.Chapters)
	}
}
