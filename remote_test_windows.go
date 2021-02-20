// +build windows

package main

import (
	"reflect"
	"testing"
)

func TestParsePackageConf(t *testing.T) {
	bits := []byte("set package_list=%package_list%\r\nlotem/rime-aoyu\r\n  lotem/rime-bopomofo-script\r\n)")
	result := []string{"lotem/rime-aoyu", "lotem/rime-bopomofo-script"}
	strs, err := parsePackageConf(bits)
	if !reflect.DeepEqual(strs, result) {
		t.Errorf("parsePackageConf failed, expect %v, got %v", result, strs)
	}
	if err != nil {
		t.Errorf("parsePackageConf failed, expected nil error, got %s", err)
	}
}
