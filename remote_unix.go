// +build linux freebsd openbsd netbsd darwin

package main

func getSep() []byte {
	return []byte("package_list=(")
}
