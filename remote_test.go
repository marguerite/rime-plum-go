package main

import (
	"reflect"
	"testing"
)

func TestParseRemotePackageConf(t *testing.T) {
	str := "lotem/rime-forge@master/lotem-packages.conf"
	result := []string{"lotem/rime-aoyu", "lotem/rime-bopomofo-script", "lotem/rime-dungfungpuo", "lotem/rime-guhuwubi", "lotem/rime-ipa", "lotem/rime-linguistic-wubi", "lotem/rime-kana", "lotem/rime-wubi98", "lotem/rime-zhengma", "lotem/rime-zhung"}
	strs, err := ParseRemotePackageConf(str)
	if !reflect.DeepEqual(strs, result) {
		t.Errorf("ParseRemotePackageConf failed, expected %v, got %v", result, strs)
	}
	if err != nil {
		t.Errorf("ParseRemotePackageConf failed, expected nil error, got %s", err)
	}
}

func TestValidateRemotePackageConf(t *testing.T) {
	str := "lotem/rime-forge@master/lotem-packages.conf"
	result := "https://github.com/lotem/rime-forge/raw/master/lotem-packages.conf"
	str1, err := validateRemotePackageConf(str)
	if str1 != result {
		t.Errorf("ValidateRemovePackageConf failed, expect result %s, got %s", result, str1)
	}
	if err != nil {
		t.Errorf("ValidateRemotePackageConf failed, expect nil error, got %s", err)
	}
}
