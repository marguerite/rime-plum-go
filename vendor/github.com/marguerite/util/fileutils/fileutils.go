package fileutils

import (
	"errors"
	"fmt"
	"github.com/marguerite/util/dirutils"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

// Touch a file or directory
func Touch(path string, isDir ...bool) error {
	// process isDir, we don't allow > 1 arguments.
	if len(isDir) > 1 {
		return errors.New("isDir is a symbol indicating whether the path is a target directory. You shall not pass two arguments.")
	}
	ok := false
	if len(isDir) == 1 {
		ok = isDir[0]
	}

	_, err := os.Stat(path)

	if err != nil {
		if os.IsNotExist(err) {
			if ok {
				err := dirutils.MkdirP(path)
				return err
			}

			dir := filepath.Dir(path)
			if dir != "." {
				err := dirutils.MkdirP(dir)
				if err != nil {
					fmt.Println("Can not create containing directory " + dir)
					return err
				}
			}
			f, err := os.Create(path)
			defer f.Close()
			if err != nil {
				if os.IsPermission(err) {
					fmt.Println("WARNING: no permission to create " + path + ", skipped...")
					return nil
				}
				return err
			}
		} else {
			return fmt.Errorf("Another unhandled non-IsNotExist PathError occurs: %s", err.Error())
		}
	}

	return nil
}

func cp(f, dst, orig string) ([]string, error) {
	// f always exists
	fi, _ := os.Stat(f)
	in, err := ioutil.ReadFile(f)
	if err != nil {
		return []string{}, err
	}

	di, err := os.Stat(dst)
	// just copy for non-existent target
	if os.IsNotExist(err) {
		err1 := ioutil.WriteFile(dst, in, fi.Mode())
		if err1 != nil {
			return []string{}, err1
		}
		return []string{dst}, nil
	} else {
		// dst here can only be file or dir because previous ReadSymlink in copy()
		if di.Mode().IsDir() {
			dst = filepath.Join(dst, filepath.Base(f))
			err1 := ioutil.WriteFile(dst, in, fi.Mode())
			if err1 != nil {
				return []string{}, err1
			}
			return []string{dst}, nil
		} else {
			err1 := ioutil.WriteFile(dst, in, fi.Mode())
			if err1 != nil {
				return []string{}, err1
			}
			if len(orig) > 0 {
				err1 = os.RemoveAll(orig)
				if err1 != nil {
					return []string{}, err1
				}
				err1 = os.Symlink(dst, orig)
				if err1 != nil {
					return []string{}, err1
				}
			}
			return []string{dst}, nil
		}
	}
	return []string{}, nil
}

func copy(f string, dst string, re []*regexp.Regexp, fn func(s, d, o string) ([]string, error)) ([]string, error) {
	fi, err := os.Stat(f)
	if err != nil {
		fmt.Println(f + " to copy does not exist, please check again")
		return []string{}, err
	}

	// file is a symlink, copy its original content
	if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
		fmt.Println(f + " is a symlink, copying the original file")
		link, err := dirutils.ReadSymlink(f)
		if err != nil {
			return []string{}, err
		}
		info, _ := os.Stat(link)
		fi = info
		f = link
	}

	orig := ""
	di, err := os.Stat(dst)
	// dst exists and is a symlink
	if !os.IsNotExist(err) && di.Mode()&os.ModeSymlink == os.ModeSymlink {
		fmt.Println(dst + " is a symlink, copy to its original file and symlink back")
		orig = dst
		link, err := dirutils.ReadSymlink(dst)
		if err != nil {
			return []string{}, err
		}
		info, _ := os.Stat(link)
		di = info
		dst = link
	}

	if fi.Mode().IsRegular() {
		files, err1 := fn(f, dst, orig)
		if err1 != nil {
			return []string{}, err1
		}
		return files, nil
	}
	if fi.Mode().IsDir() {
		if !di.Mode().IsDir() {
			return []string{}, fmt.Errorf("Source is a directory, destination should be directory too")
		}

		files, err1 := dirutils.Ls(f)
		if err1 != nil {
			return []string{}, err1
		}

		res := []string{}
		for _, v := range files {
			ok := false
			if len(re) > 0 {
				for _, r := range re {
					if r.MatchString(v) {
						ok = true
						break
					}
				}
			} else {
				ok = true
			}

			if !ok {
				continue
			}

			copyed, err1 := fn(v, dst, orig)
			if err1 != nil {
				return []string{}, err1
			}
			for _, c := range copyed {
				res = append(res, c)
			}
		}
		return res, nil
	}
	return []string{}, fmt.Errorf("Unknown FileMode %v of source", fi)
}

// Copy like Linux's cp command, copy a file/dirctory to another place.
func Copy(src, dst string) error {
	f, r := dirutils.ParseWildcard(src)
	_, err := copy(f, dst, r, cp)
	return err
}

// HasPrefixSuffixInGroup if a string's prefix/suffix matches one in group
// b trigger's prefix match
func HasPrefixSuffixInGroup(s string, group []string, b bool) bool {
	prefix := "(?i)"
	suffix := ""
	if b {
		prefix += "^"
	} else {
		suffix += "$"
	}

	for _, v := range group {
		re := regexp.MustCompile(prefix + regexp.QuoteMeta(v) + suffix)
		if re.MatchString(s) {
			return true
		}
	}
	return false
}
