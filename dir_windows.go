// +build windows

package main

import (
	"golang.org/x/sys/windows/registry"
	"os"
	"path/filepath"
	"fmt"
)

func getRimeDir() {
	key, err := registry.OpenKey(registry.CURRENT_USER, 'SOFTWARE\Rime\Weasel', registry.QUERY_VALUE)
	if err != nil {
		panic(err)
	}
	defer k.Close()

	rimeUserDir, _, err := k.GetStringValue("RimeUserDir")
	if err != nil {
		panic(err)
	}

	if len(rimeUserDir) > 0 {
		return rimeUserDir
	}

	dirAppData := os.getenv("APPDATA")
	if len(dirAppData) == 0 {
		fmt.Printf("could not find APPDATA dir.\r\n")
		os.Exit(1)
	}

	return filepath.Join(dirAppData, "Rime")
}
