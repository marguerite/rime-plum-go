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

func appendByte(byteArr ...[]byte) []byte {
	var idx int
	for _, b := range byteArr {
		idx += len(b)
	}

	res := make([]byte, idx)

	tmpIdx := 0
	for _, b := range byteArr {
		for _, bt := range b {
			res[tmpIdx] = bt
			tmpIdx += 1
		}
        }

	return res
}

func splitFiles(str string) []string {
	if !strings.Contains(str, " ") {
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
