package main

import (
	"errors"
	"testing"
)

// Chưa tick xác nhận quyền dùng tài liệu thì phần Go từ chối render, kể cả khi
// giao diện bị lách.
func TestStartRender_BatBuocXacNhanQuyen(t *testing.T) {
	a := &App{}
	if _, err := a.StartRender(BookSettings{Path: "sach.docx"}); !errors.Is(err, ErrRightsNotConfirmed) {
		t.Fatalf("muốn ErrRightsNotConfirmed, có %v", err)
	}
}
