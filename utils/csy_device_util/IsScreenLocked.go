package csy_device_util

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"unsafe"

	"github.com/godbus/dbus/v5"
	"golang.org/x/sys/windows"
)

// IsScreenLocked 返回当前屏幕是否被锁定（即处于锁屏状态）。
func IsScreenLocked() (bool, error) {
	switch runtime.GOOS {
	case "windows":
		return isLockedWindows()
	case "darwin":
		return isLockedDarwin()
	case "linux":
		return isLockedLinux()
	default:
		return false, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func isLockedWindows() (bool, error) {
	user32 := windows.NewLazySystemDLL("user32.dll")
	openInputDesktop := user32.NewProc("OpenInputDesktop")
	getUserObjectInformation := user32.NewProc("GetUserObjectInformationW")
	closeDesktop := user32.NewProc("CloseDesktop")

	const (
		DF_ALLOWOTHERACCOUNTHOOK = 0x0001
		UOI_NAME                 = 2
	)

	// 尝试打开当前接受输入交互的桌面
	hDesktop, _, err := openInputDesktop.Call(0, 0, DF_ALLOWOTHERACCOUNTHOOK)
	if err != nil {
		return false, err
	}
	if hDesktop == 0 {
		// 如果无法打开输入桌面（通常发生在 Winlogon 桌面或者权限隔离时），可视为锁定状态
		return true, nil
	}
	defer closeDesktop.Call(hDesktop)

	// 获取桌面名称
	var buf [256]uint16
	var needed uint32
	ret, _, _ := getUserObjectInformation.Call(
		hDesktop,
		uintptr(UOI_NAME),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)*2),
		uintptr(unsafe.Pointer(&needed)),
	)

	if ret == 0 {
		return false, fmt.Errorf("GetUserObjectInformationW failed")
	}

	deskName := windows.UTF16ToString(buf[:])
	// 正常解锁状态下，桌面名称为 "Default"
	// 锁屏状态下，桌面名称通常为 "Winlogon" 或无法访问 OpenInputDesktop
	return !strings.EqualFold(deskName, "Default"), nil
}

// ------------------- macOS 实现 -------------------
// 使用 CoreGraphics 的 C API 查询会话字典
// 需要 CGO，且编译时必须能链接到 CoreGraphics 框架
/*
#cgo LDFLAGS: -framework CoreGraphics -framework CoreFoundation
#include <CoreGraphics/CoreGraphics.h>
#include <CoreFoundation/CoreFoundation.h>

// 辅助函数：从 CFDictionary 中取布尔值
int getLockedFromDict(CFDictionaryRef dict) {
    CFBooleanRef locked = (CFBooleanRef)CFDictionaryGetValue(dict, CFSTR("CGSSessionScreenIsLocked"));
    if (locked && CFGetTypeID(locked) == CFBooleanGetTypeID()) {
        return CFBooleanGetValue(locked);
    }
    return 0;
}
*/
func isLockedDarwin() (bool, error) {
	out, err := exec.Command("ioreg", "-r", "-d", "1", "-c", "IOConsoleUsers").Output()
	if err != nil {
		return false, err
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, `"CGSSessionScreenIsLocked"`) {
			if strings.Contains(line, "= 1") {
				return true, nil
			}
			return false, nil
		}
	}
	return false, fmt.Errorf("could not find lock status in ioreg output")
}

// ------------------- Linux 实现 -------------------
// 通过 DBus 查询桌面环境的锁屏状态
// 依次尝试常见的 DBus 接口（GNOME / Freedesktop）
func isLockedLinux() (bool, error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return false, fmt.Errorf("failed to connect to session bus: %v", err)
	}
	defer conn.Close()

	// 尝试调用多个接口
	interfaces := []struct {
		dest, path, iface, method string
	}{
		{"org.gnome.ScreenSaver", "/org/gnome/ScreenSaver", "org.gnome.ScreenSaver", "GetActive"},
		{"org.freedesktop.ScreenSaver", "/org/freedesktop/ScreenSaver", "org.freedesktop.ScreenSaver", "GetActive"},
	}

	var lastErr error
	for _, iface := range interfaces {
		obj := conn.Object(iface.dest, dbus.ObjectPath(iface.path))
		var active bool
		err := obj.Call(iface.iface+"."+iface.method, 0).Store(&active)
		if err == nil {
			return active, nil
		}
		lastErr = err
	}
	return false, fmt.Errorf("all DBus calls failed, last error: %v", lastErr)
}
