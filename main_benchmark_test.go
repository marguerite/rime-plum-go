package main

import (
	"testing"
	"slices"
)

func BenchmarkSliceIndex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		slices.Index(TestSlice, "blueberry")
	}
}

func BenchmarkLoopAndSwitch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, v := range TestSlice {
			switch v {
			case "blueberry":
				return
			}
		}
	}
}
