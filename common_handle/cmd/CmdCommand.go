package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/front-ck996/csy/utils/csy_charset_util"
)

type Cmd struct {
	Origin         *exec.Cmd
	PrintCmd       bool
	RunBeforeCdDir string
	CmdC           bool // 仅 Windows 有效，是否使用 cmd /C
	StreamStdinCB  func(text string)
	StreamStderrCB func(text string)
	Stdin          io.WriteCloser
	Stdout         io.ReadCloser
	StdoutText     string
	Stderr         io.ReadCloser
	StderrText     string
	Proxy          string
	Timeout        time.Duration // 超时，0 表示无超时
}

func NewCMD() Cmd {
	return Cmd{
		PrintCmd: true,
		CmdC:     true,
	}
}

func (c *Cmd) isWindows() bool {
	return runtime.GOOS == "windows"
}

func (c *Cmd) convCharset(text string) string {
	// 简单处理：Windows 下若检测到非 UTF-8 则转 GBK
	if c.isWindows() {
		// 检测是否为 UTF-8，若不是则转码
		if !strings.Contains(text, "\xef\xbb\xbf") { // 简化，可用 chardet 但性能开销大
			// 使用 csy_charset_util 转换
			utf8, _ := csy_charset_util.GbkToUtf8([]byte(text))
			return string(utf8)
		}
	}
	return text
}

func (c *Cmd) Run(inputCmd []string) (string, error) {
	if len(inputCmd) == 0 {
		return "", fmt.Errorf("empty command")
	}

	var cmd *exec.Cmd

	// 判断是否为旧风格（整个命令字符串），如果是则用 shell 执行
	if len(inputCmd) == 1 && strings.Contains(inputCmd[0], " ") {
		// 使用 shell 执行
		if c.isWindows() {
			if c.CmdC {
				cmd = exec.Command("cmd", "/C", inputCmd[0])
			} else {
				cmd = exec.Command(inputCmd[0]) // 不推荐
			}
		} else {
			cmd = exec.Command("sh", "-c", inputCmd[0])
		}
	} else {
		// 新风格：命令+参数
		cmd = exec.Command(inputCmd[0], inputCmd[1:]...)
		// 如果是 Windows 且用户要求 cmd /C，但新风格下不需要，因为直接调用可执行文件
		// 但如果用户需要 cmd 内置命令（如 dir），则仍需 shell
		// 这里简单处理：如果 CmdC 为 true 且是 Windows，且第一个元素不是可执行文件路径，则用 cmd /C
		if c.isWindows() && c.CmdC {
			// 判断是否为 cmd 内置命令（如 dir, echo）
			// 简单判断：如果 inputCmd[0] 不包含路径分隔符且不是 .exe/.com 结尾，则视为内置
			// 这里简化为总是使用 cmd /C（为了兼容旧行为）
			cmd = exec.Command("cmd", "/C", strings.Join(inputCmd, " "))
		}
	}

	if c.RunBeforeCdDir != "" {
		cmd.Dir = c.RunBeforeCdDir
	}

	if c.Proxy != "" {
		env := os.Environ()
		env = append(env, "HTTP_PROXY="+c.Proxy, "HTTPS_PROXY="+c.Proxy)
		cmd.Env = env
	}

	if c.PrintCmd {
		fmt.Println("========>>")
		fmt.Println(cmd.String())
		fmt.Println("<<=======")
	}

	c.Origin = cmd

	// 管道
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}
	c.Stdin = stdin

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	c.Stdout = stdout

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}
	c.Stderr = stderr

	var wg sync.WaitGroup
	wg.Add(2)

	// 读取 stdout
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(c.Stdout)
		for scanner.Scan() {
			text := c.convCharset(scanner.Text())
			if c.StreamStdinCB != nil {
				c.StreamStdinCB(text)
			}
			c.StdoutText += text + "\n"
		}
	}()

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(c.Stderr)
		for scanner.Scan() {
			text := c.convCharset(scanner.Text())
			if c.StreamStderrCB != nil {
				c.StreamStderrCB(text)
			}
			c.StderrText += text + "\n"
		}
	}()

	err = cmd.Start()
	if err != nil {
		return "", err
	}

	// 处理超时
	if c.Timeout > 0 {
		timer := time.AfterFunc(c.Timeout, func() {
			cmd.Process.Kill()
		})
		defer timer.Stop()
	}

	err = cmd.Wait()
	// 关闭输入流（可能在 Wait 后关闭）
	c.Stdin.Close()
	// 等待所有输出读取完成
	wg.Wait()

	if err != nil {
		// 判断是否超时
		if c.Timeout > 0 && err.Error() == "signal: killed" {
			return c.StderrText, fmt.Errorf("command timed out")
		}
		return c.StderrText, err
	}
	return c.StdoutText, nil
}

// Close 关闭 stdin
func (c *Cmd) Close() error {
	return c.Stdin.Close()
}

func (c *Cmd) Exit() {
	if c.Origin != nil && c.Origin.Process != nil {
		c.Origin.Process.Kill()
	}
}
