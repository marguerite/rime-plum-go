package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/marguerite/go-stdlib/dir"
	"github.com/marguerite/go-stdlib/exec"
	"github.com/marguerite/go-stdlib/slice"
)

func buildPackages() {
	files, _ := dir.Ls(RIME_DIR, true, true)
	if ok, _ := slice.Contains(files, "essay.txt"); ok {
		minEssay(filepath.Join(RIME_DIR, "essay.txt"))
	}
	if ok, _ := slice.Contains(files, "luna_pinyin.dict.yaml"); ok {
		minLuna(filepath.Join(RIME_DIR, "luna_pinyin.dict.yaml"))
	}
	schemas, _ := dir.Glob("*.schema.yaml", RIME_DIR)
	for _, v := range schemas {
		minSchema(v)
	}
	overrideDefaultYaml(schemas)
	_, _, err := exec.Exec3("/usr/bin/rime_deployer", "--build", RIME_DIR)
	if err != nil {
		fmt.Printf("failed to run command: /usr/bin/rime_deployer --build %s\n", RIME_DIR)
		os.Exit(1)
	}
	err = os.RemoveAll(filepath.Join(RIME_DIR, "user.yaml"))
	if err != nil {
		fmt.Printf("failed to remove %s\n", filepath.Join(RIME_DIR, "user.yaml"))
		os.Exit(1)
	}
}

func overrideDefaultYaml(schemas []string) {
	m := []string{}
	for _, v := range schemas {
		s := strings.TrimRight(filepath.Base(v), ".schema.yaml")
		m = append(m, "- schema: "+s)
	}

	d := filepath.Join(filepath.Dir(schemas[0]), "default.yaml")
	f, _ := ioutil.ReadFile(d)
	re := regexp.MustCompile(`^config_version:\s+\'(.*)\'$`)
	scanner := bufio.NewScanner(strings.NewReader(string(f)))
	list := ""
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "- schema:") {
			for _, i := range m {
				if strings.Contains(line, i) {
					list += line + "\n"
					continue
				}
			}
		} else {
			if re.MatchString(line) {
				list += "config_version: '" + re.FindStringSubmatch(line)[1] + ".minimal'\n"
			} else {
				list += line + "\n"
			}
		}
	}

	err := ioutil.WriteFile(d, []byte(list), 0644)
	if err != nil {
		fmt.Printf("failed to write %s\n", d)
		os.Exit(1)
	}
}

func minEssay(s string) {
	f, _ := ioutil.ReadFile(s)
	essay := ""
	scanner := bufio.NewScanner(strings.NewReader(string(f)))
	re := regexp.MustCompile(`^.*\s+(.*)$`)
	for scanner.Scan() {
		line := scanner.Text()
		if re.MatchString(line) {
			idx := re.FindStringSubmatch(line)[1]
			i, _ := strconv.Atoi(idx)
			if i >= 500 {
				essay += line + "\n"
			}
		}
	}
	err := ioutil.WriteFile(s, []byte(essay), 0644)
	if err != nil {
		fmt.Printf("failed to write %s\n", s)
		os.Exit(1)
	}
}

func minLuna(s string) {
	f, _ := ioutil.ReadFile(s)
	scanner := bufio.NewScanner(strings.NewReader(string(f)))
	luna := ""
	re := regexp.MustCompile(`^version:\s+\"(.*)\"$`)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#以下") {
			break
		}
		if re.MatchString(line) {
			luna += "version:\t\"" + re.FindStringSubmatch(line)[1] + ".minimal\"\n"
		} else {
			luna += line + "\n"
		}
	}
	err := ioutil.WriteFile(s, []byte(luna), 0644)
	if err != nil {
		fmt.Printf("failed to write file %s\n", s)
		os.Exit(1)
	}
}

func minSchema(s string) {
	f, _ := ioutil.ReadFile(s)
	scanner := bufio.NewScanner(strings.NewReader(string(f)))
	schema := ""
	re := regexp.MustCompile(`(\s+version:\s+)\"(.*)\"$`)
	for scanner.Scan() {
		line := scanner.Text()
		if re.MatchString(line) {
			m := re.FindStringSubmatch(line)
			schema += m[1] + "\"" + m[2] + ".minimal\"\n"
			continue
		}
		if strings.Contains(line, "- stroke") {
			schema += "#" + line + "\n"
			continue
		}
		if strings.Contains(line, "- reverse_lookup_translator") {
			schema += "#" + line + "\n"
		}
		schema += line + "\n"
	}
	err := ioutil.WriteFile(s, []byte(schema), 0644)
	if err != nil {
		fmt.Printf("failed to write file %s\n", s)
		os.Exit(1)
	}
}
