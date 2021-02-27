package main

import (
	"testing"
)

func TestGenRx(t *testing.T) {
	pkg := Package{"emoji", "rime", "https://github.com", "", "", true, "customize", map[string]string{"schema": "terra_pinyin"}, ""}
	str := "# Rx: https://github.com/rime/emoji:customize:schema=terra_pinyin {\n"
	rx := genRx(pkg)
	if rx != str {
		t.Errorf("genRx test failed, expect %s, got %s", str, rx)
	}
}

func TestParsePackageFromPatchContent(t *testing.T) {
	pkg := Package{"emoji", "rime", "https://github.com", "master", "https://github.com/rime/emoji", true, "customize", map[string]string{"schema": "terra_pinyin"}, ""}
	rx := "# Rx: https://github.com/rime/emoji:customize:schema=terra_pinyin {\n"

	pkg1 := parsePackageFromPatchContent(str2bytes(rx))
	if !pkg.equal(pkg1) {
		t.Errorf("parsePackageFromPatchContent test failed, expected %v, got %v", pkg, pkg1)
	}
}

func TestCountSpace(t *testing.T) {
	b := str2bytes("    Rx: ")
	n := countSpace(b)
	if n != 4 {
		t.Errorf("countSpace failed, expected 4, got %d", n)
	}
}

func TestParseCustomYamlWithNoPatchLine(t *testing.T) {
	pkg := Package{}
	i, j := parseCustomYaml([]byte{}, pkg)
	if i != -1 || j != -1 {
		t.Errorf("parseCustomYaml failed, expeced (-1, -1), got (%d, %d)", i, j)
	}
}

func TestParseCustomYamlWithJustPatchLine(t *testing.T) {
	pkg := Package{}
	i, j := parseCustomYaml(str2bytes("__patch:\n"), pkg)
	if i != 0 || j != -1 {
		t.Errorf("parseCustomYaml failed, expeced (0, -1), got (%d, %d)", i, j)
	}
}

func TestParseCustomYamlWithHashMarkAsEndingByte(t *testing.T) {
	pkg := Package{}
	i, j := parseCustomYaml(str2bytes("__patch:\n#"), pkg)
	if i != 0 || j != -1 {
		t.Errorf("parseCustomYaml failed, expeced (0, -1), got (%d, %d)", i, j)
	}
}

func TestParseCustomYamlWithNoSpaceFollowingHashMark(t *testing.T) {
	pkg := Package{}
	i, j := parseCustomYaml(str2bytes("__patch:\n#Hello"), pkg)
	if i != 0 || j != -1 {
		t.Errorf("parseCustomYaml failed, expeced (0, -1), got (%d, %d)", i, j)
	}
}

func TestParseCustomYamlWithNoRx(t *testing.T) {
	pkg := Package{}
	i, j := parseCustomYaml(str2bytes("__patch:\n# "), pkg)
	if i != 0 || j != -1 {
		t.Errorf("parseCustomYaml failed, expeced (0, -1), got (%d, %d)", i, j)
	}
}

func TestParseCustomYamlWithOtherStringButRx(t *testing.T) {
	pkg := Package{}
	i, j := parseCustomYaml(str2bytes("__patch:\n# Recipe"), pkg)
	if i != 0 || j != -1 {
		t.Errorf("parseCustomYaml failed, expeced (0, -1), got (%d, %d)", i, j)
	}
}

func TestParseCustomYamlWithRxButNoSecondHashMark(t *testing.T) {
	pkg := Package{}
	i, j := parseCustomYaml(str2bytes("__patch:\n# Rx: Hello"), pkg)
	if i != 0 || j != -1 {
		t.Errorf("parseCustomYaml failed, expeced (0, -1), got (%d, %d)", i, j)
	}
}

func TestParseCustomYamlWithTheSecondHashMarkAsEndingByte(t *testing.T) {
	pkg := Package{}
	i, j := parseCustomYaml(str2bytes("__patch:\n# Rx: #"), pkg)
	if i != 0 || j != -1 {
		t.Errorf("parseCustomYaml failed, expeced (0, -1), got (%d, %d)", i, j)
	}
}

func TestParseCustomYamlWithNoSpaceFollowingSecondHashMark(t *testing.T) {
	pkg := Package{}
	i, j := parseCustomYaml(str2bytes("__patch:\n# Rx: #Hello"), pkg)
	if i != 0 || j != -1 {
		t.Errorf("parseCustomYaml failed, expeced (0, -1), got (%d, %d)", i, j)
	}
}

func TestParseCustomYamlWithSecondHashMarkAndSpaceAsEndingByte(t *testing.T) {
	pkg := Package{}
	i, j := parseCustomYaml(str2bytes("__patch:\n# Rx: # "), pkg)
	if i != 0 || j != -1 {
		t.Errorf("parseCustomYaml failed, expeced (0, -1), got (%d, %d)", i, j)
	}
}

func TestParseCustomYamlWithNoClosingCurlyBraket(t *testing.T) {
	pkg := Package{}
	i, j := parseCustomYaml(str2bytes("__patch:\n# Rx: # Hello"), pkg)
	if i != 0 || j != -1 {
		t.Errorf("parseCustomYaml failed, expeced (0, -1), got (%d, %d)", i, j)
	}
}

func TestParseCustomYamlWithNoClosingCurlyBraket1(t *testing.T) {
	pkg := Package{}
	i, j := parseCustomYaml(str2bytes("__patch:\n# Rx: # H"), pkg)
	if i != 0 || j != -1 {
		t.Errorf("parseCustomYaml failed, expeced (0, -1), got (%d, %d)", i, j)
	}
}

func TestParseCustomYamlWithCurlyBraketAsEndingByte(t *testing.T) {
	pkg := Package{"emoji", "rime", "https://github.com", "master", "https://github.com/rime/emoji", true, "customize", map[string]string{"schema": "terra_pinyin"}, ""}
	b := str2bytes("__patch:\n# Rx: https://github.com/rime/emoji:customize:schema=terra_pinyin {\n# }")
	i, j := parseCustomYaml(b, pkg)
	if i != 9 || j != 79 {
		t.Errorf("parseCustomYaml failed, expeced (9, 79), got (%d, %d)", i, j)
	}
}

func TestParseCustomYaml(t *testing.T) {
	pkg := Package{"emoji", "rime", "https://github.com", "master", "https://github.com/rime/emoji", true, "customize", map[string]string{"schema": "terra_pinyin"}, ""}
	i, j := parseCustomYaml(str2bytes("__patch:\n# Rx: https://github.com/rime/emoji:customize:schema=terra_pinyin {\n# }\n"), pkg)
	if i != 9 || j != 80 {
		t.Errorf("parseCustomYaml failed, expeced (9, 80), got (%d, %d)", i, j)
	}
}

func TestParseCustomYaml1(t *testing.T) {
	pkg := Package{"emoji", "rime", "https://github.com", "master", "https://github.com/rime/emoji", true, "customize", map[string]string{"schema": "terra_pinyin"}, ""}
	i, j := parseCustomYaml(str2bytes("__patch:\n# Rx: https://github.com/rime/emoji:customize:schema=terra_pinyin {\n# }\n#"), pkg)
	if i != 9 || j != 80 {
		t.Errorf("parseCustomYaml failed, expeced (9, 80), got (%d, %d)", i, j)
	}
}

func TestParseCustomYamlWithNoMatchingRecipe(t *testing.T) {
	pkg := Package{"emoji1", "rime", "https://github.com", "master", "https://github.com/rime/emoji1", true, "customize", map[string]string{"schema": "terra_pinyin"}, ""}
	i, j := parseCustomYaml(str2bytes("__patch:\n# Rx: https://github.com/rime/emoji:customize:schema=terra_pinyin {\n# }\n#"), pkg)
	if i != 71 || j != -1 {
		t.Errorf("parseCustomYaml failed, expeced (71, -1), got (%d, %d)", i, j)
	}
}

func TestLoopPatchString(t *testing.T) {
	str := "luna_pinyin"
	str1 := loopPatch(str, 0)
	if str1 != " luna_pinyin\n" {
		t.Errorf("loopPatch test failed, expected \" luna_pinyin\n\", got \"%s\"", str1)
	}
}

func TestLoopPatchMap(t *testing.T) {
	str := "schema: luna_pinyin"
	str1 := loopPatch(str, 0)
	if str1 != " schema: luna_pinyin\n" {
		t.Errorf("loopPatch test failed, expected \" schema: luna_pinyin\n\", got \"%s\"", str1)
	}
}

func TestLoopPatchSlice(t *testing.T) {
	str := []string{"luna_pinyin"}
	str1 := loopPatch(str, 0)
	if str1 != "\t luna_pinyin\n" {
		t.Errorf("loopPatch test failed, expected \"\t luna_pinyin\n\", got \"%s\"", str1)
	}
}
