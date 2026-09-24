package bookmaker

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func writeTemp(t *testing.T, name string, data []byte) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestParseDocx_TuChoiFileCoKhoa(t *testing.T) {
	data := append(append([]byte{}, cfbMagic...), make([]byte, 512)...)
	data = append(data, encryptionInfoUTF16...)
	if _, err := ParseDocx(writeTemp(t, "khoa.docx", data)); !errors.Is(err, ErrProtectedFile) {
		t.Fatalf("muốn ErrProtectedFile, có %v", err)
	}
}

func TestParseDocx_BaoRoDocCu(t *testing.T) {
	data := append(append([]byte{}, cfbMagic...), make([]byte, 512)...)
	if _, err := ParseDocx(writeTemp(t, "cu.doc", data)); !errors.Is(err, ErrLegacyDoc) {
		t.Fatalf("muốn ErrLegacyDoc, có %v", err)
	}
}
