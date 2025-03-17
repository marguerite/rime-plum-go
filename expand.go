// expand variable
// https://www.gnu.org/software/bash/manual/html_node/Shell-Parameter-Expansion.html#Shell-Parameter-Expansion
// currently only supports ${parameter%word} and ${parameter:-word}, you can implement others here
package main

import (
	"bytes"
	"fmt"
	"os"
)

func expandVar(b []byte, args map[string]string) []byte {
	b = b[2 : len(b)-1]
	var param []byte

	for i, v := range b {
		switch v {
		case ':':
			switch b[i+1] {
			case '-':
				word := b[i+2:]
				if args == nil {
					return word
				}
				val, ok := args[string(param)]
				if !ok {
					return word
				}
				return []byte(val)
			case '=', '?', '+':
				fmt.Printf("not implemented yet %s\n", b)
				os.Exit(1)
			default:
				// offset:length
				fmt.Printf("not implemented %s\n", b)
				os.Exit(1)
			}
		case '%':
			if b[i+1] == '%' {
				fmt.Printf("not implemented %s\n", b)
				os.Exit(1)
			}
			// %%
			word := b[i+2:]
			val := args[string(param)]
			j := bytes.Index([]byte(val), word)
			if j < 0 {
				return []byte(val)
			}
			return []byte(val[:j])
		case '#', '/', '^', ',', '@':
			fmt.Printf("not implemented %s\n", b)
			os.Exit(1)
		default:
			param = append(param, v)
		}
	}
	return b
}
