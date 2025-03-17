package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	RIME_DIR       = os.Getenv("RIME_DIR")
	RIME_FRONTEND  = os.Getenv("RIME_FRONTEND")
	PRESET_SCHEMAS = []string{"bopomofo", "cangjie", "custom", "essay", "luna-pinyin", "prelude", "stroke", "terra-pinyin"}
	EXTRA_SCHEMAS  = []string{"array", "cantonese", "combo-pinyin", "double-pinyin", "emoji", "essay-simp", "ipa", "middle-chinese", "pinyin-simp", "quick", "scj", "soutzoe", "stenotype", "wubi", "wugniu", "emoji-cantonese"}
	LEN_PRESET     = len(PRESET_SCHEMAS)
	LEN_EXTRA      = len(EXTRA_SCHEMAS)
	LEN_ALL        = LEN_PRESET + LEN_EXTRA
	CLIENT         = &http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
)

func main() {
	args := os.Args

	if len(args) == 1 {
		fmt.Println("please input rime configurations, as ':preset' or '(https://github.com)/(lotem)/rime-forge(@master|/raw/master)/(lotem-packages.conf)'. the quoted parts can be skipped if you're using github.com/rime prefix. 'plum' itself with root permission will update this binary.")
		os.Exit(0)
	}

	if len(args) == 2 && args[1] == "plum" {
		// update ourselves
		fmt.Println("Not Implemented Yet")
		os.Exit(0)
	}

	// help section

	var build, ui bool

	for i, v := range args[1:] {
		switch v {
		case "--build":
			build = true
		case "--select":
			ui = true
		}
		if build || ui {
			args = append(args[:i], args[i+1:]...)
		}
		if build && ui {
			break
		}
	}

	pkgs := make(PackageSets, 0, len(args[1:]))
	var i int

	for _, v := range args[1:] {
		pkg, err := NewPackageSet(v)
		if err != nil {
			panic(err)
		}
		pkgs = append(pkgs, pkg)
		i++
	}

	if len(RIME_FRONTEND) == 0 {
		RIME_FRONTEND = GetRimeFrontend()
	}

	if len(RIME_DIR) == 0 {
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
