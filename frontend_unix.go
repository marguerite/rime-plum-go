// +build linux freebsd openbsd netbsd

package main

import (
	"os"
	"path/filepath"
)

func getRimeFrontend() string {
	im := os.Getenv("GTK_IM_MODULE")
	if len(im) == 0 {
		for _, v := range []string{"fcitx5", "fcitx", "ibus"} {
			if stat, err := os.Stat(filepath.Join("/usr/bin", v)); err == nil && stat.Mode()&0111 != 0 {
				return v
			}
		}
	}

	return im
}
