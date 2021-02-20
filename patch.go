package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

func (r Recipe) patch(pkg Package) {
	for k, v := range r.PatchFiles {
		fmt.Printf("Patching %s\n", k)
		content := genPatchContent(pkg, v)
		filename := filepath.Join(RIME_DIR, k)
		str := "__patch:\n" + content
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			ioutil.WriteFile(filename, []byte(str), 0644)
		} else {
			b, err := ioutil.ReadFile(filename)
			if err != nil {
				panic(err)
			}
			fmt.Printf("Updating patch...\n")
			begin, end := parseCustomYaml(b, pkg)

			if begin < 0 {
				// append "__patch:\n" and our patch
				for _, v := range []byte(str) {
					b = append(b, v)
				}
				ioutil.WriteFile(filename, b, 0644)
				continue
			}
			if end < 0 {
				// append our patch to the end of "__patch:\n" section
				tmp := b[:begin]
				for _, v := range []byte(content) {
					tmp = append(tmp, v)
				}
				if begin < len(b)-1 {
					for _, v := range b[begin+1:] {
						tmp = append(tmp, v)
					}
				}
				ioutil.WriteFile(filename, tmp, 0644)
				continue
			}
			// replace our old patch
			tmp := b[:begin]
			for _, v := range []byte(content) {
				tmp = append(tmp, v)
			}
			for _, v := range b[end:] {
				tmp = append(tmp, v)
			}
			ioutil.WriteFile(filename, tmp, 0644)
		}
	}
}

// parseCustomYaml parse the existing .custom.yaml contains "__patch:\n" and old patch
// returns a starting byte and an ending byte
// both < 0 means no "__patch:\n" found
// the ending byte < 0 means we find "__patch:\n" located at starting byte,
// but we can not find any old patch applied
func parseCustomYaml(b []byte, pkg Package) (int, int) {
	idx := bytes.Index(b, []byte("__patch:\n"))
	if idx < 0 {
		return -1, -1
	}
	b2 := b[idx+9:] // 9: the length of "__patch:\n" plus 1

	var begin, end, idx1 int
	var bs [][]int

	for {
		i := bytes.Index(b2, []byte("#"))
		if i < 0 {
			break
		}
		b2 = b2[i+1:]

		var j int
		for {
			if b2[0] != ' ' {
				break
			}
			b2 = b2[1:]
			j++
		}

		if !bytes.HasPrefix(b2, []byte("Rx:")) {
			if b2[0] == '}' {
				end = begin + idx1 + i + j + 4             // 4: is a magic word I have no time to figure out why
				bs = append(bs, []int{begin + 9, end + 9}) // add 9 because original "b" has "__patch:\n"
				begin = 0
				idx1 = 0
				b2 = b2[j+1:]
				continue
			}
			continue
		}
		begin = end + i
		idx1 += j
	}

	for _, v := range bs {
		if pkg.equal(parsePackageFromPatchContent(b[v[0]:v[1]])) {
			return v[0], v[1]
		}
	}
	return bs[len(bs)-1][1], -1
}

func parsePackageFromPatchContent(b []byte) Package {
	idx := bytes.Index(b, []byte("Rx: "))
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

func genRx(pkg Package) string {
	var str string
	str += "# Rx: " + pkg.Host + "/" + pkg.User + "/" + pkg.Repo + ":" + pkg.Rx
	if pkg.RxOptions != nil {
		str += ":"
		for k, v := range pkg.RxOptions {
			str += k + "=" + v + ","
		}
	}
	if str[len(str)-1] == ',' {
		str = str[:len(str)-1]
	}
	str += " {\n"
	return str
}

func genPatchContent(pkg Package, patch Patch) string {
	str := genRx(pkg)
	str += loopPatch(patch, 0)
	str += "# }\n"
	return str
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
