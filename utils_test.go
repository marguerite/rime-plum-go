package main

import (
	"reflect"
	"testing"
)

func TestSplit(t *testing.T) {
	str := "key=value,key1=value1"
	result := []string{"key=value", "key1=value1"}
	arr := split(str, ",", 2)
	if !reflect.DeepEqual(arr, result) {
		t.Errorf("split failed, expect %v, got %v", result, arr)
	}
}

func TestSplitFiles(t *testing.T) {
	str := "# emoji_suggestion.yaml # opencc/*.* *.txt *.json"
	arr := splitFiles(str)
	if arr[0] != "*.txt" || arr[1] != "*.json" {
		t.Errorf("splitFiles failed, expect %v, got %v", []string{"*.txt", "*.json"}, arr)
	}
}
