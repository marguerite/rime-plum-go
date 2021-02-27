// +build linux freebsd openbsd netbsd darwin

package main

func getSep() []byte {
	return str2bytes("package_list=(")
}
