package main

import (
	"reflect"
	"testing"
)

func TestNewPackage(t *testing.T) {
	testArgs := []string{"jyutping",
		"lotem/rime-zhung",
		"lotem/rime-zhung@master",
	}

	testResults := []Package{{"rime-jyutping", "rime", "https://github.com", "master", "https://github.com/rime/rime-jyutping", true, "", nil, ""},
		{"rime-zhung", "lotem", "https://github.com", "master", "https://github.com/lotem/rime-zhung", true, "", nil, ""},
		{"rime-zhung", "lotem", "https://github.com", "master", "https://github.com/lotem/rime-zhung", true, "", nil, ""},
	}

	for i, v := range testArgs {
		pkg := NewPackage(v)
		if !pkg.equal(testResults[i]) {
			t.Errorf("NewPackage failed, expected %v, got %v", testResults[i], pkg)
		}
	}
}

func TestParseRx(t *testing.T) {
	arr := []string{"@master:custom:key=value,key1=value1", "rime-prelude:custom:key=value", "rime-prelude:custom", "rime-prelude"}
	result := [][]string{{"@master", "custom"}, {"rime-prelude", "custom"}, {"rime-prelude", "custom"}, {"rime-prelude", ""}}
	result1 := []map[string]string{{"key": "value", "key1": "value1"}, {"key": "value"}, nil, nil}
	for i, v := range arr {
		x, y, z := parseRx(v)
		if x != result[i][0] || y != result[i][1] || !reflect.DeepEqual(z, result1[i]) {
			t.Errorf("parseRx failed, expected %s %s %v, got %s %s %v", result[i][0], result[i][1], result1[i], x, y, z)
		}
	}
}
