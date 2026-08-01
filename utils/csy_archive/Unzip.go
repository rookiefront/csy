package csy_archive

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
)

// Unzip 解压 ZIP 文件，自动处理 GBK 编码的文件名
func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	for _, f := range r.File {
		// ----- 处理文件名编码 -----
		fileName := f.Name
		if f.NonUTF8 { // 标志位表示非 UTF‑8，按 GBK 解码
			decoded, err := simplifiedchinese.GBK.NewDecoder().String(fileName)
			if err == nil {
				fileName = decoded
			}
			// 解码失败则保留原始名称（虽然可能乱码，但避免中断）
		}

		// 构建安全的解压路径
		fpath := filepath.Join(dest, fileName)

		// 防止 Zip Slip 攻击（路径穿越）
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("非法文件路径: %s", fpath)
		}

		// 处理目录
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, f.Mode()); err != nil {
				return err
			}
			continue
		}

		// 处理普通文件
		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		defer outFile.Close()

		if _, err := io.Copy(outFile, rc); err != nil {
			return err
		}
	}
	return nil
}
