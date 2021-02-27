// +build windows

package main

func getSep() []byte {
	return str2bytes("package_list=%package_list%")
}
