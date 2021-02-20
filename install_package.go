package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/marguerite/go-stdlib/dir"
	"github.com/marguerite/go-stdlib/fileutils"
)

func (pkg Package) install() {
	if len(pkg.Rx) > 0 {
		r := filepath.Join(pkg.WorkingDirectory, pkg.Rx+".recipe.yaml")
		if _, err := os.Stat(r); os.IsNotExist(err) {
			fmt.Printf("no recipe %s found, typo?\n", pkg.Rx)
			os.Exit(1)
		}
		b, err := ioutil.ReadFile(r)
		if err != nil {
			fmt.Printf("failed to open file %s: %s\n", r, err)
			os.Exit(1)
		}
		recipe, err := NewRecipe(b, pkg.RxOptions)
		if err != nil {
			fmt.Printf("failed to parse recipe: %s\n", err)
			os.Exit(1)
		}
		checkRecipe(recipe, pkg)
		recipe.download(pkg)
		recipe.install(pkg)
		recipe.patch(pkg)
		return
	}

	if _, err := os.Stat(filepath.Join(pkg.WorkingDirectory, "recipe.yaml")); !os.IsNotExist(err) {
		r := filepath.Join(pkg.WorkingDirectory, "recipe.yaml")
		b, err := ioutil.ReadFile(r)
		if err != nil {
			fmt.Printf("failed to read file %s: %s\n", r, err)
			os.Exit(1)
		}
		recipe, err := NewRecipe(b, pkg.RxOptions)
		if err != nil {
			fmt.Printf("failed to parse recipe: %s\n", err)
			os.Exit(1)
		}
		checkRecipe(recipe, pkg)
		recipe.download(pkg)
		recipe.install(pkg)
		recipe.patch(pkg)
		return
	}

	installFilesFromDir(pkg.WorkingDirectory)
}

func checkRecipe(r Recipe, pkg Package) {
	if len(pkg.Rx) == 0 {
		return
	}
	if pkg.Rx != r.Recipe.Rx {
		fmt.Printf("invalid recipe: %s does not match file name %s\n", r.Recipe.Rx, pkg.Rx)
		os.Exit(1)
	}
}

func installFilesFromDir(d string) {
	pattern := [][]string{[]string{"*.yaml", "{custom,recipe}.yaml"}, []string{"*.txt", "opencc/"}, []string{"*.gram"}, []string{"opencc/*.*", ".{json,ocd,txt}"}}

	var files []string
	for _, v := range pattern {
		var matches []string
		var err error
		if len(v) > 1 {
			matches, err = dir.Glob(v[0], d, v[1])
		} else {
			matches, err = dir.Glob(v[0], d)
		}
		if err != nil {
			fmt.Printf("can not find qualified files in %s\n", d)
			os.Exit(1)
		}
		for _, m := range matches {
			files = append(files, m)
		}
	}

	for _, v := range files {
		err := fileutils.Copy(v, RIME_DIR)
		if err != nil {
			fmt.Printf("failed to copy %s to %s\n", v, RIME_DIR)
			os.Exit(1)
		}
	}
}
