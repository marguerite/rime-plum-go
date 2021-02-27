package main

import (
	"bytes"
	"testing"
)

func BenchmarkGenRx(b *testing.B) {
	pkg := Package{"emoji", "rime", "https://github.com", "", "", true, "customize", map[string]string{"schema": "terra_pinyin"}, ""}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		genRx(pkg)
	}
}

func genRx1(pkg Package) string {
	var str string
	str += "# Rx: " + pkg.Host + "/" + pkg.User + "/" + pkg.Repo + ":" + pkg.Rx
	if pkg.RxOptions != nil {
		str += ":"
		for k, v := range pkg.RxOptions {
			str += k + "=" + v + ","
		}
		str = str[:len(str)-1]
	}
	str += " {\n"
	return str
}

func BenchmarkGenRx1(b *testing.B) {
	pkg := Package{"emoji", "rime", "https://github.com", "", "", true, "customize", map[string]string{"schema": "cantonese"}, ""}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		genRx1(pkg)
	}
}

func BenchmarkParsePackageFromPatchContent(b *testing.B) {
	rx := str2bytes("# Rx: https://github.com/rime/emoji:customize:schema=terra_pinyin {\n")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parsePackageFromPatchContent(rx)
	}
}

func parsePackageFromPatchContent1(b []byte) Package {
	idx := bytes.Index(b, str2bytes("Rx: "))
	b = b[idx+4:]
	b1 := make([]byte, 0, 80)
	for {
		if b[0] == ' ' {
			break
		}
		b1 = append(b1, b[0])
		b = b[1:]
	}
	return NewPackage(string(b1))
}

func BenchmarkParsePackageFromPatchContent1(b *testing.B) {
	rx := str2bytes("# Rx: https://github.com/rime/emoji:customize:schema=terra_pinyin {\n")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parsePackageFromPatchContent1(rx)
	}
}
