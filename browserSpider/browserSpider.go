package browserSpider

import (
	"context"
	"github.com/chromedp/chromedp"
	"github.com/front-ck996/csy"
	"log"
	"os"
	"time"
)

type BrowserHandle struct {
	Ctx      context.Context
	Headless bool
	Close    context.CancelFunc
	UA       string
	TempDir  string
	NoDel    bool
}

type BrowserHandleInit struct {
	TempDir string
	NoDel   bool
	Opts    []chromedp.ExecAllocatorOption
}

func New(init BrowserHandleInit) BrowserHandle {
	c := BrowserHandle{
		TempDir: init.TempDir,
		NoDel:   init.NoDel,
	}
	if c.UA == "" {
		c.UA = "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.87 Safari/537.36"
	}
	if c.TempDir == "" {
		c.TempDir = "c:\\click-temp\\"
	}
	if c.NoDel {
		dir, _ := csy.NewFile().IsDir(c.TempDir)
		if !dir {
			os.MkdirAll(c.TempDir, os.ModePerm)
		}
	} else {
		os.RemoveAll(c.TempDir)
		os.MkdirAll(c.TempDir, os.ModePerm)
	}
	dir := c.TempDir

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent(c.UA),
		chromedp.DisableGPU,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("headless", c.Headless),
		//chromedp.Flag("start-maximized", true),
		//chromedp.Flag("ignore-certificate-errors", true),
		//chromedp.Flag("incognito", true),
		chromedp.Flag("window-size", "1380,900"),
		chromedp.UserDataDir(dir),
		chromedp.Flag("disable-infobars", true),
		chromedp.Flag("excludeSwitches", `['enable-automation','foo']`),
		//chromedp.Flag("useAutomationExtension", `False`),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("remote-debugging-port", "9222"),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("enable-blink-features", "IdleDetection"),
	//--enable-blink-features=IdleDetection
	)
	opts = append(opts, chromedp.ExecPath(GetBrowserExe()))

	opts = append(opts, init.Opts...)

	var allocCtx context.Context
	var ctxxx context.Context
	allocCtx, c.Close = chromedp.NewExecAllocator(context.Background(), opts...)

	// also set up a custom logger
	ctxxx, c.Close = chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	//程序允许得总耗时
	c.Ctx, c.Close = context.WithTimeout(ctxxx, time.Hour*90000)
	//page.AddScriptToEvaluateOnNewDocument()

	chromedp.Run(c.Ctx, DelWebdriver())
	return c
}

// 关闭浏览器
func (c *BrowserHandle) Off() {
	c.Close()
}

func GetBrowserExe() string {
	chromelist := []string{
		`C:\Users\123\Desktop\Chrome-bin\chrome.exe`,
		`C:\Users\Administrator\Desktop\chrome\Chrome-bin\chrome.exe`,
		`C:\Users\Administrator\Desktop\chrome-win\chrome.exe`,
		`C:\Users\Administrator\AppData\Local\Google\Chrome\Application\chrome.exe`,
		`C:\Users\Administrator\AppData\Local\Google\Chrome\Bin\chrome.exe`,
		`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
		`C:\Program Files\Google\Chrome\Application\chrome.exe`,
	}
	for _, chrome_v := range chromelist {
		if csy.NewFile().IsFile(chrome_v) {
			return chrome_v
		}
	}
	return ""
}
