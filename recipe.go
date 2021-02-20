package main

import (
	"bytes"
	"fmt"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// Patch the content of Patch
type Patch interface{}

// Recipe configuration to standardly customize rime
type Recipe struct {
	Recipe struct {
		Rx          string   `yaml:"Rx"`
		Args        []string `yaml:"args,omitempty"`
		Description string   `yaml:"description"`
	} `yaml:"recipe"`
	DownloadFiles string           `yaml:"download_files,omitempty"`
	InstallFiles  string           `yaml:"install_files,omitempty"`
	PatchFiles    map[string]Patch `yaml:"patch_files"`
}

// NewRecipe initialize a recipe struct
func NewRecipe(b []byte, options map[string]string) (Recipe, error) {
	var r, r1 Recipe
	err := yaml.Unmarshal(b, &r)
	if err != nil {
		fmt.Printf("failed to unmarshal yaml %s: %s\n", b, err)
		return r, err
	}

	if len(r.Recipe.Args) > 0 {
		args := parseArg(r.Recipe.Args, options)
		b1 := replaceVar(b, args)
		err = yaml.Unmarshal(b1, &r1)
		if err != nil {
			fmt.Printf("failed to unmarshal yaml %s: %s\n", b1, err)
			return r1, err
		}
		return r1, nil
	}
	return r, nil
}

func replaceVar(b []byte, args map[string]string) []byte {
	var b1 []byte
	for {
		i := bytes.Index(b, []byte("${"))
		if i < 0 {
			b1 = appendbyte(b1, b)
			break
		}
		j := bytes.IndexByte(b[i:], '}')
		v := b[i : i+j+1]
		b1 = appendbyte(b1, b[:i])
		b1 = appendbyte(b1, expandVar(v, args))
		if i+j+1 == len(b) {
			break
		}
		b = b[i+j+1:]
	}
	return b1
}

func parseArg(args []string, options map[string]string) map[string]string {
	args1 := make(map[string]string, len(args))
	for _, v := range args {
		i := strings.Index(v, "=")
		if i < 0 {
			args1[v] = ""
		} else {
			arr := split(v, "=", 2)
			args1[arr[0]] = arr[1]
		}
	}

	for k, v := range options {
		if _, ok := args1[k]; ok {
			args1[k] = v
		} else {
			fmt.Printf("no such key %s, you may provide a wrong parameter\n", k)
		}
	}

	return args1
}
