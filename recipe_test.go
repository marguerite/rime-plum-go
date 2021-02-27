package main

import "testing"

func TestReplaceVar(t *testing.T) {
	b := str2bytes("schema: ${schema:-luna_pinyin}")
	result := "schema: terra_pinyin"
	args := map[string]string{"schema": "terra_pinyin"}
	b1 := replaceVar(b, args)
	if string(b1) != result {
		t.Errorf("replaceVar failed, expected %s, got %s", result, string(b1))
	}
}

func TestParseArg(t *testing.T) {
	args := []string{"schema=luna_pinyin"}
	options := map[string]string{"schema": "terra_pinyin"}
	args1 := parseArg(args, options)
	val, ok := args1["schema"]
	if !ok || val != "terra_pinyin" {
		t.Errorf("parseArg failed, expected terra_pinyin, got %s", val)
	}
}
