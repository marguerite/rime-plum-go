// +build linux freebsd openbsd netbsd

package main

import (
	"os"
	"path/filepath"
)

func getRimeFrontend() string {
	im := os.Getenv("GTK_IM_MODULE")
	if im == "ibus" {
		return im
	}
	for _, v := range []string{"fcitx5", "fcitx"} {
		if stat, err := os.Stat(filepath.Join("/usr/bin", v)); err == nil && stat.Mode()&0111 != 0 {
			return v
		}
	}

	return im
}
