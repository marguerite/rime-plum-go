package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"unicode"
)

func (r Recipe) patch(pkg Package) {
	for k, v := range r.PatchFiles {
		fmt.Printf("Patching %s\n", k)
		content := genPatchContent(pkg, v)
		filename := filepath.Join(RIME_DIR, k)
		b := str2bytes("__patch:\n" + content)
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			ioutil.WriteFile(filename, b, 0644)
		} else {
			b1, err := ioutil.ReadFile(filename)
			if err != nil {
				panic(err)
			}
			fmt.Printf("Updating patch...\n")
			begin, end := parseCustomYaml(b1, pkg)

			if begin < 0 {
				// append "__patch:\n" and our patch
				b1 = appendbyte(b1, b)
				ioutil.WriteFile(filename, b1, 0644)
				continue
			}
			if end < 0 {
				// append our patch to the end of "__patch:\n" section
				tmp := b1[:begin+9]
				tmp = appendbyte(tmp, str2bytes(content))
				if begin+8 < len(b1)-1 {
					tmp = appendbyte(tmp, b[begin+9:])
				}
				ioutil.WriteFile(filename, tmp, 0644)
				continue
			}
			// replace our old patch
			tmp := b1[:begin]
			tmp = appendbyte(tmp, str2bytes(content))
			tmp = appendbyte(tmp, b1[end+1:])
			ioutil.WriteFile(filename, tmp, 0644)
		}
	}
}

// parseCustomYaml parse the existing .custom.yaml that may contains "__patch:\n" and an old patch
// returns a starting byte and an ending byte
// both < 0 means no "__patch:\n" found
// the ending byte < 0 means we find "__patch:\n" located at starting byte,
// but we can not find any old patch applied
func parseCustomYaml(b []byte, pkg Package) (int, int) {
	idx := bytes.Index(b, str2bytes("__patch:\n"))
	if idx < 0 {
		return -1, -1
	}
	b2 := b[idx+9:] // 9: the length of "__patch:\n" plus 1

	// step: count how many steps we passed
	var step int
	var bs [][]int

	for {
		i := bytes.Index(b2, str2bytes("#"))
		if i < 0 {
			// no patched recipe found in .custom.yaml
			break
		}

		// if i is the last byte, break
		if i == len(b2)-1 {
			break
		}

		j := countSpace(b2[i+1:])

		// no space found, it means we may see other kind of comment, just step forward
		if j == 0 {
			step += i + 1
			b2 = b2[i+1:]
			continue
		}

		// if i + j is the last byte, break
		if i+j == len(b2)-1 {
			break
		}

		if bytes.HasPrefix(b2[i+j+1:], str2bytes("Rx:")) {
			// m count starts from "Rx:"
			//                      ^ position 0 here, which is i+j+1
			m := bytes.Index(b2[i+j+1:], str2bytes("#"))
			if m < 0 {
				step += i + j + 1
				b2 = b2[i+j+1:]
			}

			// "# abcdefg# h"
			//  0101234567
			//  0123456789
			//  ij       mn
			//  n = 10
			//  n = i + j + 1 + m + 1
			//    = 0 + 1 + 1 + 7 + 1
			if i+j+1+m == len(b2)-1 {
				break
			}
			n := countSpace(b2[i+j+1+m+1:])
			if n == 0 {
				step += i + j + 1 + m + 1
				b2 = b2[i+j+1+m+1:]
				continue
			}

			// "# abcdefg# }"
			//  0101234567
			//  01234567890
			//  ij       mn
			//  tmp = 10
			//  tmp = i + j + 1 + m + n
			//      = 0 + 1 + 1 + 7 + 1
			tmp := i + j + 1 + m + n
			if tmp == len(b2)-1 {
				break
			}

			if b2[tmp+1] == '}' {
				if tmp+1 == len(b2)-1 {
					bs = append(bs, []int{step + i, step + tmp + 1})
					break
				}
				// the following "\n"
				bs = append(bs, []int{step + i, step + tmp + 2})
				if tmp+2 == len(b2)-1 {
					break
				}
				step += tmp + 3
				b2 = b2[tmp+3:]
			} else {
				if tmp+1 == len(b2)-1 {
					break
				}
				step += tmp + 1
				b2 = b2[tmp+1:]
			}
		} else {
			// add step and process
			step += i + j + 1
			b2 = b2[i+j+1:]
		}
	}

	for _, v := range bs {
		if pkg.equal(parsePackageFromPatchContent(b[v[0]+9 : v[1]+10])) {
			return v[0] + 9, v[1] + 9
		}
	}

	if len(bs) > 0 {
		return bs[len(bs)-1][1], -1
	}
	return idx, -1
}

// countSpace find out how many white space
func countSpace(b []byte) int {
	var i int
	for {
		if !unicode.IsSpace(rune(b[0])) {
			break
		}
		i++
		if len(b) == 1 {
			break
		}
		b = b[1:]
	}
	return i
}

func parsePackageFromPatchContent(b []byte) Package {
	idx := bytes.Index(b, str2bytes("Rx: "))
	b = b[idx+4:]
	idx1 := bytes.Index(b, str2bytes(" "))

	return NewPackage(string(b[:idx1]))
}

func genRx(pkg Package) string {
	var b bytes.Buffer
	b.WriteString("# Rx: ")
	b.WriteString(pkg.Host)
	b.WriteByte('/')
	b.WriteString(pkg.User)
	b.WriteByte('/')
	b.WriteString(pkg.Repo)
	b.WriteByte(':')
	b.WriteString(pkg.Rx)
	if pkg.RxOptions != nil {
		b.WriteByte(':')
		for k, v := range pkg.RxOptions {
			b.WriteString(k)
			b.WriteByte('=')
			b.WriteString(v)
			b.WriteByte(',')
		}
		b.Truncate(b.Len() - 1)
	}
	b.WriteString(" {\n")
	return b.String()
}

func genPatchContent(pkg Package, patch Patch) string {
	var b bytes.Buffer
	b.WriteString(genRx(pkg))
	b.WriteString(loopPatch(patch, 0))
	b.WriteString("# }\n")
	return b.String()
}

func loopPatch(v interface{}, idx int) string {
	var str string
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Map:
		iter := rv.MapRange()
		for iter.Next() {
			var suffix string
			kd := reflect.ValueOf(iter.Value().Interface()).Kind()
			if kd == reflect.Map || kd == reflect.Slice {
				suffix = "\n"
			}
			val, _ := iter.Key().Interface().(string)
			str += strings.Repeat("\t", idx) + "- " + val + ":" + suffix
			if reflect.ValueOf(iter.Value().Interface()).Kind() != reflect.String || rv.Len() > 1 {
				str += loopPatch(iter.Value().Interface(), idx+1)
			} else {
				str += loopPatch(iter.Value().Interface(), 0)
			}
		}
	case reflect.Slice:
		for i := 0; i < rv.Len(); i++ {
			str += loopPatch(rv.Index(i).Interface(), idx+1)
		}
	case reflect.String:
		str += strings.Repeat("\t", idx) + " " + rv.String() + "\n"
	default:
		fmt.Printf("undecided type of patch content: %v, %v", rv, rv.Kind())
		os.Exit(1)
	}
	return str
}
