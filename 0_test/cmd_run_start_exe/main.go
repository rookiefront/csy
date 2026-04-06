package main

import (
	"fmt"

	"github.com/front-ck996/csy/common_handle/cmd"
)

func main() {
	cmd := cmd.Cmd{}
	cmd.StreamStderrCB = func(text string) {

	}
	cmd.StreamStdinCB = func(text string) {
		fmt.Println(text)
		cmd.InputText("213")
	}
	run, err := cmd.Run([]string{
		"G:\\code\\my\\csy\\0_test\\cmd_run_start_exe\\ask.exe",
		//"ping", "-t", "baidu.com",
	})
	fmt.Println(run, err)

	fmt.Println(cmd)
}
