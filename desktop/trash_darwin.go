package main

import (
	"os"
	"path/filepath"
)

// moveToSystemTrash dời vào ~/.Trash (Thùng rác của Finder).
func moveToSystemTrash(path string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	_, err = moveAside(path, filepath.Join(home, ".Trash"))
	return err
}
