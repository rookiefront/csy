package cmd_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/front-ck996/csy/common_handle/cmd"
)

func Test_01(t *testing.T) {
	os.RemoveAll("element-plus")
	cmd := cmd.NewCMD()
	cmd.StreamStdinCB = func(text string) {
		fmt.Println("stdin", text)
	}
	cmd.StreamStderrCB = func(text string) {
		fmt.Println("stderr", text)
	}
	cmd.Run([]string{
		"git", "clone", "--progres", "git@github.com:element-plus/element-plus.git",
	})
}
