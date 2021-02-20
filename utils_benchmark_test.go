package main

import (
	"strings"
	"testing"
)

func BenchmarkSplit(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	str := "key=value,key1=value1"
	for i := 0; i < b.N; i++ {
		split(str, ",", 2)
	}
}

func BenchmarkSplit1(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	str := "key=value,key1=value1"
	for i := 0; i < b.N; i++ {
		strings.Split(str, ",")
	}
}
