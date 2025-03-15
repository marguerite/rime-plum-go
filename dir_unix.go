// +build linux freebsd openbsd netbsd darwin

package main

import (
	"os"
	"path/filepath"
)

func getRimeDir() string {
	home, err := os.UserHomeDir()
	if err != nil || len(home) == 0 {
		panic("no User Home directory found")
	}

	switch RIME_FRONTEND {
	case "fcitx5":
		return filepath.Join(home, ".local/share/fcitx5/rime")
	case "fcitx":
		return filepath.Join(home, ".config/fcitx/rime")
	case "ibus":
		return filepath.Join(home, ".config/ibus/rime")
	case "squirrel":
		return filepath.Join(home, "Library/Rime")
	default:
		panic("No such RIME_FRONTEND, available ones: fcitx5, fcitx, ibus, squirrel, weasel.")
	}

}
