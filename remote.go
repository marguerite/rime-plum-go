package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

//validateRemotePackageConf validate remote rime package.conf
func validateRemotePackageConf(str string) (string, error) {
	// fillup host
	if !strings.HasPrefix(str, "http") {
		str = "https://github.com/" + str
	}
	// replace the @master bits
	idx := strings.Index(str, "@")

	if idx > 0 {
		bits := make([]byte, len(str)-idx)
		var j int

		for i := idx; i < len(str); i++ {
			if str[i] == '/' {
				break
			}
			if i == len(str)-1 {
				break
			}
			bits = append(bits, str[i])
			j++
		}
		branch := string(bits[len(bits)-j:])
		str = strings.Replace(str, branch, "/raw/"+branch[1:], 1)
	}

	_, err := url.Parse(str)
	if err != nil {
		return str, err
	}

	return str, nil
}

//parsePackageConf parse rime package.conf
func parsePackageConf(bits []byte) ([]string, error) {
	strs := make([]string, bytes.Count(bits, []byte{'\n'}))
	//[]byte("#!/bin/bash\n\npackage_list=(\n\tlotem/rime-aoyu\n  lotem/rime-bopomofo-script\n)")
	sep := getSep()
	idx := bytes.Index(bits, sep)

	if idx < 0 {
		return strs, errors.New("not a valid package.conf")
	}

	tmp := make([]byte, 80)
	var j, m int

	for i := idx + len(sep); i < len(bits); i++ {
		switch bits[i] {
		case '\n':
			if j > 0 {
				strs = append(strs, string(tmp[len(tmp)-j:]))
				m++
				tmp = []byte{}
				j = 0
			}
		case '\t', ' ', '\r', ')':
			continue
		default:
			tmp = append(tmp, bits[i])
			j++
		}
	}

	return strs[len(strs)-m:], nil
}

//ParseRemotePackageConf parse remote rime package.conf
func ParseRemotePackageConf(str string) ([]string, error) {
	var strs []string
	url, err := validateRemotePackageConf(str)
	if err != nil {
		return strs, err
	}

	resp, err := CLIENT.Get(url)
	if err != nil {
		return strs, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return strs, fmt.Errorf("%s not fetched, status %d", url, resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return strs, err
	}

	strs, err = parsePackageConf(b)
	if err != nil {
		return strs, err
	}

	return strs, nil
}
