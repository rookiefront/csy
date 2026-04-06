package csy_runtime_util

import (
	"os"
)

func GetMainDir() string {
	dir, _ := os.Getwd()
	return dir
}
