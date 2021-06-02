// +build windows

package main

import (
	"golang.org/x/sys/windows/registry"
	"path/filepath"
	"strings"
	"runtime"
)

func getRimeDeployer() (string, string) {
	var s string
	if strings.HasSuffix(runtime.GOARCH, "64") {
		s = "SOFTWARE\\Wow6432Node\\Rime\\Weasel"
	} else {
		s = "SOFTWARE\\Rime\\Weasel"
	}
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, s, registry.QUERY_VALUE)
	if err != nil {
		panic(err)
	}
	defer func() {
		err := k.Close()
		if err != nil {
			panic(err)
		}
	}()
	val, _, err := k.GetStringValue("WeaselRoot")
	if err != nil {
		panic(err)
	}
	val1, err := registry.ExpandString(val)
	if err != nil {
		panic(err)
	}
	return filepath.Join(val1, "WeaselDeployer.exe"), "WeaselDeployer.exe not found, please install weasel."
}
