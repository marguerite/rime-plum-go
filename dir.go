package main

import (
	"os"
	"path/filepath"
)

// GetRimeDir get Rime Data Dir
func GetRimeDir() string {
	home, err := os.UserHomeDir()
	if err != nil || len(home) == 0 {
		panic("no HOME directory found")
	}

	switch RIME_FRONTEND {
	case "fcitx5":
		return filepath.Join(home, ".config/fcitx5/rime")
	case "fcitx":
		return filepath.Join(home, ".config/fcitx/rime")
	case "ibus":
		return filepath.Join(home, ".config/ibus/rime")
	case "weasel":
		appdata := os.Getenv("APPDATA")
		if len(appdata) == 0 {
			panic("no APPDATA directory found")
		}
		return filepath.Join(appdata, "Rime")
	case "squirrel":
		return filepath.Join(home, "Library/Rime")
	default:
		panic("unknown frontend")
	}
}
