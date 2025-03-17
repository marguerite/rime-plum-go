// plum has these kinds of command line options:
// 1. :preset default package sets
// 2. jyutping one default package with https://github.com/rime/rime-jyutping skipped
// 3. lotem/rime-zhung one user package with https://github.com/ skipped
// 4. lotem/rime-zhung@master package with branch specified
// 5. lotem/rime-forge@master/lotem-packages.conf a packages.conf file
// 6. https://github.com/lotem/rime-forge/raw/master/lotem-packages.conf vanilla url
// 7. "plum" update plum itself
// rime_dir="$HOME/.config/fcitx/rime" and rime_frontend=fcitx-rime global environment variables
// --select a ncurses-like UI to select packages to install
package main

import (
	"fmt"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/storage/memory"
)

func getAllSchemas() []string {
	arr := make([]string, LEN_PRESET+LEN_EXTRA)
	for i := 0; i < LEN_PRESET; i++ {
		arr[i] = PRESET_SCHEMAS[i]
	}
	for i := 0; i < LEN_EXTRA; i++ {
		arr[LEN_PRESET+i] = EXTRA_SCHEMAS[i]
	}
	return arr
}

// PackageSets rime package sets specified via command line. eg: ":preset jyutping", are two package sets, :preset can be expanded to a package set with many packages, while jyutping is a package set with one package
type PackageSets []PackageSet

func (r *PackageSets) toggle(idx int, ck bool) {
	var i int
	for m := range *r {
		for n := range (*r)[m].Packages {
			if i == idx {
				(*r)[m].Packages[n].Install = ck
				break
			}
			i++
		}
	}
}

// Packages rime packages
type PackageSet struct {
	Packages []Package
	Preset   bool
	Raw      string
	File     string
}

// NewPackageSet initialize new PackageSet
func NewPackageSet(str string) (PackageSet, error) {
	var pkgs PackageSet
	pkgs.Raw = str

	if len(str) == 0 {
		return pkgs, nil
	}

	if str[0] == ':' {
		pkgs.Preset = true
		str_arr := make([]string, LEN_ALL)
		switch str[1:] {
		case "preset":
			for i := 0; i < LEN_PRESET; i++ {
				str_arr[i] = PRESET_SCHEMAS[i]
			}
			str_arr = str_arr[:LEN_PRESET]
		case "extra":
			for i := 0; i < LEN_EXTRA; i++ {
				str_arr[i] = EXTRA_SCHEMAS[i]
			}
			str_arr = str_arr[:LEN_EXTRA]
		case "all":
			for i, v := range getAllSchemas() {
				str_arr[i] = v
			}
		default:
			return pkgs, fmt.Errorf("not a valid preset set %s", str)
		}
		packages := make([]Package, 0, len(str_arr))
		for i, v := range str_arr {
			packages[i] = NewPackage(v)
		}
		pkgs.Packages = packages
		return pkgs, nil
	}

	if strings.HasSuffix(str, ".conf") {
		// two kinds of file format, one is the raw URL, the other is the @master style
		pkgs.File = str
		str_arr, err := ParseRemotePackageConf(str)
		if err != nil {
			return pkgs, err
		}
		packages := make([]Package, 0, len(str_arr))
		for i, v := range str_arr {
			packages[i] = NewPackage(v)
		}
		pkgs.Packages = packages
		return pkgs, nil
	}

	pkgs.Packages = []Package{NewPackage(str)}

	return pkgs, nil
}

// Package rime package
type Package struct {
	Repo             string            // jyutping or terra-pinyin?
	User             string            // rime or lotem or maguerite?
	Host             string            // github.com?
	Branch           string            // master?
	URL              string            // to vanilla url
	Install          bool              // whether to install recipe
	Rx               string            // custom.yaml
	RxOptions        map[string]string // key=value
	WorkingDirectory string
}

func (pkg Package) equal(pkg1 Package) bool {
	if pkg.URL != pkg1.URL ||
		pkg.Branch != pkg1.Branch ||
		pkg.Rx != pkg1.Rx {
		return false
	}
	for k, v := range pkg.RxOptions {
		if val, ok := pkg1.RxOptions[k]; ok {
			if v != val {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

// fillmissing fill up default host, user and branch
func (pkg *Package) normalize() {
	if len(pkg.Host) == 0 {
		pkg.Host = "https://github.com"
	}
	if len(pkg.User) == 0 {
		pkg.User = "rime"
	}

	// can't use filepath because it will eat double "/" to "/", thus go-git will treat that URL as ssh
	pkg.URL = pkg.Host + "/" + pkg.User + "/" + pkg.Repo
	
	rem := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{pkg.URL},
	})

	refs, _ := rem.List(&git.ListOptions{
		PeelingOption: git.AppendPeeled,
	})

	var branches []string

	for _, ref := range refs {
		if ref.Name().IsBranch() {
			branches = append(branches, strings.TrimPrefix(ref.Name().String(), "refs/heads/"))
		}
	}

	for _, branch := range branches {
		if branch == "master" {
			pkg.Branch = "master"
			return
		}
	}

	pkg.Branch = "main"
}

// NewPackage initialize a rime package
func NewPackage(str string) Package {
	var pkg Package
	pkg.Install = true

	f := func(s string, pkg *Package) {
		j := strings.Index(s, "@")
		if j < 0 {
			pkg.Repo, pkg.Rx, pkg.RxOptions = parseRx(s)
		} else {
			pkg.Repo = s[:j]
			pkg.Branch, pkg.Rx, pkg.RxOptions = parseRx(s[j+1:])
		}
	}

	i := strings.Count(str, "/")

	if i == 0 {
		// jyutping
		f(str, &pkg)
		pkg.Repo = "rime-" + pkg.Repo
		pkg.normalize()
		return pkg
	}

	arr := split(str, "/", i+1)

	switch i {
	// lotem/rime-zhung and lotem/rime-zhung@master
	case 1:
		pkg.User = arr[0]
		f(arr[1], &pkg)
	case 2:
		// github.com/lotem/rime-forge@master:recipe#schema=luna_pinyin
		pkg.Host = arr[0]
		pkg.User = arr[1]
		f(arr[2], &pkg)
	case 4:
		// https://github.com/lotem/rime-forge@master:recipe#schema=luna_pinyin
		pkg.Host = strings.Join(arr[:3], "/")
		pkg.User = arr[3]
		f(arr[4], &pkg)
	default:
		// vanilla https://github.com/lotem/rime-forge/raw/master:recipe#schema=luna_pinyin
		pkg.Host = strings.Join(arr[:3], "/")
		pkg.User = arr[3]
		pkg.Repo = arr[4]
		j := strings.Index(arr[6], ":")
		if j < 0 {
			pkg.Branch = arr[6]
		} else {
			pkg.Branch, pkg.Rx, pkg.RxOptions = parseRx(arr[6])
		}
	}
	pkg.normalize()
	return pkg
}

// parseRx parse branch/repo, Rx and RxOptions
func parseRx(str string) (string, string, map[string]string) {
	// rime-forge:recipe:key=value
	idx := strings.Index(str, ":")

	if idx < 0 {
		return str, "", nil
	}

	branch := str[:idx]
	str = str[idx+1:]

	idx1 := strings.Index(str, ":")

	if idx1 < 0 {
		return branch, str, nil
	}

	rx := str[:idx1]

	options := str[idx1+1:]

	i := strings.Count(options, ",")

	if i == 0 {
		arr1 := split(options, "=", 2)
		return branch, rx, map[string]string{arr1[0]: arr1[1]}
	}

	arr2 := split(options, ",", i+1)
	m := make(map[string]string, len(arr2))
	for _, j := range arr2 {
		arr3 := split(j, "=", 2)
		m[arr3[0]] = arr3[1]
	}
	return branch, rx, m
}
