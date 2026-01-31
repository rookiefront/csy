package csy

import (
	"bufio"
	"fmt"
	"github.com/saintfish/chardet"
	"io"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// Cmd 封装常用的操作系统命令的函数
type Cmd struct {
	Origin *exec.Cmd
	// 打印执行的命令
	PrintCmd bool
	// 执行命令前，进入这个目录
	RunBeforeCdDir string
	CmdC           bool
	// 执行过程中输出的流
	StreamStdinCB func(text string)
	// 执行过程中错误的流
	StreamStderrCB func(text string)
	Stdin          io.WriteCloser
	Stdout         io.ReadCloser
	StdoutText     string
	Stderr         io.ReadCloser
	StderrText     string
}

func NewCMD() Cmd {
	return Cmd{
		PrintCmd: true,
		CmdC:     true,
	}
}
func (c *Cmd) convCharset(text string) string {

	detector := chardet.NewTextDetector()
	r, err := detector.DetectBest([]byte(text))
	if r.Charset == "UTF-8" && err == nil {
		return text
	}
	if c.isWindows() {
		utf8, _ := GbkToUtf8([]byte(text))
		return string(utf8)
	}
	return text
}
func (c *Cmd) isWindows() bool {
	return runtime.GOOS == "windows"
}

func (c *Cmd) InputText(input string) error {
	// 可以接受在系统过程中的输入，暂时不需要
	//scanner := bufio.NewScanner(os.Stdin)
	_, err := fmt.Fprintln(c.Stdin, input)
	return err
}

func (c *Cmd) Close() error {
	return c.Stdin.Close()
}
func (c *Cmd) Exit() {
	c.Stdin.Close()
	c.Origin.Process.Kill()
	//SIGINT 通常用于中断进程，
	//SIGTERM 通常用于正常终止进程，
	//SIGQUIT 通常用于优雅退出进程。
	//signal.Notify()
	//err := c.Origin.Process.Signal(syscall.SIGQUIT)
	//// 退出失败直接杀死进程
	//if err != nil {
	//	//err :=
	//	//if err != nil {
	//	//	fmt.Println("强制杀死进程失败", err)
	//	//}
	//	return
	//}
}

func (c *Cmd) Run(inputCmd []string) (string, error) {
	var cmdStr []string
	cmd := exec.Command("")

	switch {
	case runtime.GOOS == "windows":
		if c.CmdC {
			cmdStr = append(cmdStr, "cmd.exe", "/C")
		}
		if c.RunBeforeCdDir != "" {
			cmdStr = append(cmdStr, filepath.VolumeName(c.RunBeforeCdDir))
			cmdStr = append(cmdStr, "cd "+strings.ReplaceAll(c.RunBeforeCdDir, "\\", "/"))
		}
		cmdStr = append(cmdStr, inputCmd...)
		if c.CmdC {
			cmd = exec.Command(cmdStr[0], []string{cmdStr[1], strings.Join(cmdStr[2:], " & ")}...)
		} else {
			cmd = exec.Command(cmdStr[0], []string{strings.Join(cmdStr[1:], " & ")}...)
		}
		break
	default:
		cmd = exec.Command(strings.Join(inputCmd, " && "))
		//cmdStr = append(cmdStr, "bash", "-c")
	}

	if c.PrintCmd {
		fmt.Println("========>>")
		fmt.Println(cmd.String())
		fmt.Println("<<=======")
	}

	c.Origin = cmd

	// 获取输入流
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}
	c.Stdin = stdin

	// 获取输出流
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	c.Stdout = stdout

	// 获取标准错误流
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}
	c.Stderr = stderr

	var wg sync.WaitGroup
	wg.Add(2)
	// 接受标准输出流
	go func() {
		wg.Done()
		scanner := bufio.NewScanner(c.Stdout)
		for scanner.Scan() {
			text := c.convCharset(scanner.Text())
			if c.StreamStdinCB != nil {
				c.StreamStdinCB(text)
			}
			c.StdoutText += text + "\n"
		}
	}()

	// 接受标准错误流
	go func() {
		wg.Done()
		scanner := bufio.NewScanner(c.Stderr)
		for scanner.Scan() {
			text := c.convCharset(scanner.Text())
			if c.StreamStderrCB != nil {
				c.StreamStderrCB(text)
			}
			c.StderrText += text + "\n"
		}
	}()

	// 启动命令
	err = cmd.Start()
	if err != nil {
		return "", err
	}
	//// 关闭输入流
	//err = stdin.Close()
	//if err != nil {
	//	return "", err
	//}
	// 等待命令执行完毕
	wg.Wait()
	err = cmd.Wait()
	if err != nil {
		return c.StderrText, err
	}
	return c.StdoutText, nil
}
