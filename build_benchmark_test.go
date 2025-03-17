package main

import (
	"testing"
	"github.com/marguerite/go-stdlib/slice"
	"slices"
)

var (
	TestSlice = []string{"apple", "banana", "orange", "blueberry", "raspberry"}
)

func BenchmarkMargueriteSliceContains(b *testing.B) {
	for i := 0; i < b.N; i++ {
		slice.Contains(TestSlice, "blueberry")
	}
}

func BenchmarkBuiltinSlices(b *testing.B) {
	for i := 0; i < b.N; i++ {
		slices.Contains(TestSlice, "blueberry")
	}
}
