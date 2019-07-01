package main

import (
	"reflect"
	"testing"
)

func TestParsePackagesConf(t *testing.T) {
	s := "#!/bin/bash\n\npackage_list+=(\n\tlotem/rime-aoyu\n\tlotem/rime-bopomofo-script\n)"
	result := []string{"lotem/rime-aoyu", "lotem/rime-bopomofo-script"}
	recipes, err := parsePackagesConf(s)
	if err != nil {
		t.Error("parsePackagesConf test failed.")
	}
	if reflect.DeepEqual(recipes, result) {
		t.Log("parsePackagesConf test passed.")
	} else {
		t.Error("parsePackagesConf test failed.")
	}
}

func TestNewRecipeUrl(t *testing.T) {
	result := RecipeURL{"github.com", "rime", "rime-terra-pinyin", "master"}
	if reflect.DeepEqual(NewRecipeURL("", "", "terra-pinyin", ""), result) {
		t.Log("NewRecipeURL test passed.")
	} else {
		t.Error("NewRecipeURL test failed.")
	}
}

func TestParseRecipeUrl(t *testing.T) {
	result := RecipeURLs{RecipeURL{"github.com", "marguerite", "terra-pinyin", "v1.0.0"},
		RecipeURL{"marguerite.su", "lotem", "terra-pinyin", "master"}}
	recipes := []string{"marguerite/terra-pinyin@v1.0.0", "marguerite.su/lotem/terra-pinyin"}
	if reflect.DeepEqual(parseRecipeUrl(recipes), result) {
		t.Log("parseRecipeUrl test passed.")
	} else {
		t.Error("parseRecipeUrl test failed.")
	}
}
