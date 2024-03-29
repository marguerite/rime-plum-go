package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	RIME_DIR      = ""
	RIME_FRONTEND = ""
	preset        = []string{"bopomofo", "cangjie", "custom", "essay", "luna-pinyin", "prelude", "stroke", "terra-pinyin"}
	extra         = []string{"array", "cantonese", "combo-pinyin", "double-pinyin", "emoji", "essay-simp", "ipa", "middle-chinese", "pinyin-simp", "quick", "scj", "soutzoe", "stenotype", "wubi", "wugniu"}
	CLIENT        = &http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
)

func main() {
	args := os.Args[1:]

	if len(args) < 1 {
		fmt.Println("please input rime configurations, as ':preset' or '(https://github.com)/(lotem)/rime-forge(@master|/raw/master)/(lotem-packages.conf)'. the quoted parts can be skipped if you're using github.com/rime prefix. 'plum' itself with root permission will update this binary.")
		os.Exit(1)
	}

	// help section

	var build, ui bool

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--build":
			args = append(args[:i], args[i+1:]...)
			build = true
		case "--select":
			args = append(args[:i], args[i+1:]...)
			ui = true
		}
	}

	if len(args) == 1 && args[0] == "plum" {
		// update ourselves
	}

	// make append quickly, len(extra) is enough for most of the cases
	pkgs := make(PackagesStr, 0, len(args)+len(extra))
	var i int

	for _, v := range args {
		pkg, err := NewPackages(v)
		if err != nil {
			panic(err)
		}
		pkgs = append(pkgs, pkg)
		i++
	}

	pkgs = pkgs[len(pkgs)-i:]

	if len(os.Getenv("rime_frontend")) > 0 {
		RIME_FRONTEND = os.Getenv("rime_frontend")
	} else {
		RIME_FRONTEND = GetRimeFrontend()
	}

	if len(os.Getenv("rime_dir")) > 0 {
		RIME_DIR = os.Getenv("rime_dir")
	} else {
		RIME_DIR = GetRimeDir()
	}

	if ui {
		// advanced console ui
		m := model{
			choices:  pkgs,
			selected: make(map[int]struct{}),
		}
		program := tea.NewProgram(m)
		if err := program.Start(); err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}
	}

	pkgs.install()

	if build {
		buildPackages()
	}
}
