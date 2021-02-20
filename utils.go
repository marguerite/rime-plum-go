package main

import (
	"strings"
)

func split(str, sep string, size int) []string {
	arr := make([]string, 0, size)
	for {
		idx := strings.Index(str, sep)
		if idx < 0 {
			arr = append(arr, str)
			break
		}
		arr = append(arr, str[:idx])
		str = str[idx+1:]
	}
	return arr
}

func appendbyte(b []byte, b1 []byte) []byte {
	b2 := make([]byte, len(b)+len(b1))
	copy(b2, b)
	var j int
	for i := len(b); i < len(b)+len(b1); i++ {
		b2[i] = b1[j]
		j++
	}
	return b2
}

func splitFiles(str string) []string {
	if strings.Index(str, " ") < 0 {
		return []string{str}
	}
	i := strings.Count(str, " ")
	// # emoji_suggestion.yaml # opencc/*.*
	arr := make([]string, 0, i+1)
	var skip bool

	for {
		j := strings.Index(str, " ")
		if j < 0 {
			if !skip {
				arr = append(arr, str)
			}
			break
		}
		if skip {
			str = str[j+1:]
			skip = false
			continue
		}
		if str[:j] == "#" {
			skip = true
		} else {
			if str[0] != '#' {
				arr = append(arr, str[:j])
			}
		}
		str = str[j+1:]
	}

	return arr
}
