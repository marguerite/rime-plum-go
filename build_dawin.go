// +build darwin

package main

func getRimeDeployer() (string, string) {
	return "/Applications/Squirrel.app/Contents/MacOS/rime_deployer", "Squirrel.app not found, please install Squirrel"
}
