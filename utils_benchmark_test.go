package main

import (
	"strings"
	"testing"
)

func BenchmarkSplit(b *testing.B) {
	b.ResetTimer()
	str := "key=value,key1=value1"
	for i := 0; i < b.N; i++ {
		split(str, ",", 2)
	}
}

func BenchmarkSplit1(b *testing.B) {
	b.ResetTimer()
	str := "key=value,key1=value1"
	for i := 0; i < b.N; i++ {
		strings.Split(str, ",")
	}
}

func BenchmarkAppendByte(b *testing.B) {
	x := str2bytes("hello world")
	y := str2bytes("have a lot of fun")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x = appendbyte(x, y)
	}
}

func BenchmarkAppendByte1(b *testing.B) {
	x := "hello world"
	y := "have a lot of fun"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x += y
	}
}

func BenchmarkStr2Byte(b *testing.B) {
	str := "have a lot of fun"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str2bytes(str)
	}
}

func BenchmarkStr2Byte1(b *testing.B) {
	str := "have a lot of fun"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = []byte(str)
	}
}
