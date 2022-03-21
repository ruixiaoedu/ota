package main

import "os"

// configPath 配置
func configFilename() string {
	dir, _ := os.Getwd()
	return dir + "/ota.ini"
}
