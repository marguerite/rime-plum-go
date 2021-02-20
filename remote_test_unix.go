// +build linux freebsd openbsd netbsd darwin

package main

import (
	"reflect"
	"testing"
)

func TestParsePackageConf(t *testing.T) {
	bits := []byte("#!/bin/bash\n\npackage_list=(\n\tlotem/rime-aoyu\n  lotem/rime-bopomofo-script\n)")
	result := []string{"lotem/rime-aoyu", "lotem/rime-bopomofo-script"}
	strs, err := parsePackageConf(bits)
	if !reflect.DeepEqual(strs, result) {
		t.Errorf("parsePackageConf failed, expect %v, got %v", result, strs)
	}
	if err != nil {
		t.Errorf("parsePackageConf failed, expected nil error, got %s", err)
	}
}
