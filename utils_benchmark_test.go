package main

import (
	"strings"
	"testing"
	"math/rand"
)

func BenchmarkSplit(b *testing.B) {
	str := "key=value,key1=value1"
	for i := 0; i < b.N; i++ {
		split(str, ",", 2)
	}
}

func BenchmarkStringsSplit(b *testing.B) {
	str := "key=value,key1=value1"
	for i := 0; i < b.N; i++ {
		strings.Split(str, ",")
	}
}

func genRandByteSlice() [][]byte {
	magicNum := 100
	var size int

	for {
		size = rand.Intn(magicNum)
		if size > 0 {
			break
		}
	}
	
	res := make([][]byte, 0, size)

	for i := 0; i < size; i++ {
		var size1 int
		for {
			size1 = rand.Intn(magicNum)
			if size1 > 0 {
				break
			}
		}

		token := make([]byte, size1)
		for {
			rand.Read(token)
			if token[0] > 0 {
				break
			}
		}
		res = append(res, token)
	}

	return res
}

var benchmarkByteData = genRandByteSlice()

func BenchmarkAppendByte(b *testing.B) {
	for i := 0; i < b.N; i++ {
		appendByte(benchmarkByteData...)
	}
}

func BenchmarkBuiltinAppend(b *testing.B) {
	tmp := make([]byte, 0)
	for i := 0; i < b.N; i++ {
		for _, v := range benchmarkByteData {
			tmp = append(tmp, v...)
		}
	}
}


