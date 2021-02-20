package main

import (
	"path/filepath"

	"github.com/marguerite/go-stdlib/fileutils"
)

func (r Recipe) install(pkg Package) {
	arr := splitFiles(r.InstallFiles)
	for _, v := range arr {
		fileutils.Copy(filepath.Join(pkg.WorkingDirectory, v), filepath.Join(RIME_DIR, v))
	}
}
