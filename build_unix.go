// +build linux freebsd openbsd netbsd

package main

func getRimeDeployer() (string, string) {
	return "/usr/bin/rime_deployer", "/usr/bin/rime_deployer not found, please run `sudo zypper in --no-recommends rime` first"
}
