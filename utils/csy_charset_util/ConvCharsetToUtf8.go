package csy_charset_util

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/axgle/mahonia"
	"github.com/saintfish/chardet"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
)

// CharsetDetector 字符集检测器（单例模式）
var defaultDetector = chardet.NewTextDetector()

// ConvCharsetToUtf8 自动检测字符集并转换为 UTF-8
func ConvCharsetToUtf8(content []byte) (string, error) {
	if len(content) == 0 {
		return "", nil
	}

	// 1. 检测字符集
	result, err := defaultDetector.DetectBest(content)
	if err != nil {
		// 检测失败时，尝试直接作为 UTF-8 处理
		return string(content), fmt.Errorf("detect charset failed, use as UTF-8: %w", err)
	}

	// 2. 根据检测结果进行转换
	charset := normalizeCharset(result.Charset)

	// 调试信息（生产环境可以去掉或改为日志）
	//fmt.Printf("Detected encoding: %s (confidence: %s, language: %s)\n",
	//	result.Charset, result.Confidence, result.Language)

	// 3. 执行转换
	converted, err := ConvertToUtf8(content, charset)
	if err != nil {
		return string(content), fmt.Errorf("convert from %s to UTF-8 failed: %w", charset, err)
	}

	return converted, nil
}

// ConvertToUtf8 将指定编码的内容转换为 UTF-8
func ConvertToUtf8(content []byte, fromCharset string) (string, error) {
	switch strings.ToLower(fromCharset) {
	case "utf-8", "utf8":
		return string(content), nil

	case "gbk", "gb18030", "gb2312":
		// 使用 golang.org/x/text 方案
		decoder := simplifiedchinese.GBK.NewDecoder()
		utf8Data, err := io.ReadAll(transform.NewReader(bytes.NewReader(content), decoder))
		if err != nil {
			// 降级到 mahonia
			return convertWithMahonia(content, fromCharset)
		}
		return string(utf8Data), nil

	case "big5":
		decoder := traditionalchinese.Big5.NewDecoder()
		utf8Data, err := io.ReadAll(transform.NewReader(bytes.NewReader(content), decoder))
		if err != nil {
			return convertWithMahonia(content, fromCharset)
		}
		return string(utf8Data), nil

	case "iso-8859-1", "latin1":
		// 简单处理：每个字节直接作为 Unicode 码点
		return latin1ToUtf8(content), nil

	default:
		// 使用 mahonia 作为后备方案
		return convertWithMahonia(content, fromCharset)
	}
}

// convertWithMahonia 使用 mahonia 库进行转换
func convertWithMahonia(content []byte, charset string) (string, error) {
	decoder := mahonia.NewDecoder(charset)
	if decoder == nil {
		return "", fmt.Errorf("unsupported charset: %s", charset)
	}
	return decoder.ConvertString(string(content)), nil
}

// latin1ToUtf8 将 Latin-1 (ISO-8859-1) 转换为 UTF-8
func latin1ToUtf8(content []byte) string {
	result := make([]rune, len(content))
	for i, b := range content {
		result[i] = rune(b)
	}
	return string(result)
}

// normalizeCharset 标准化字符集名称
func normalizeCharset(charset string) string {
	normalized := strings.ToLower(strings.TrimSpace(charset))

	// 常见别名映射
	aliases := map[string]string{
		"gb18030":    "gbk",
		"gb2312":     "gbk",
		"hz-gb-2312": "gbk",
		"big-5":      "big5",
		"utf8":       "utf-8",
		"ibm866":     "windows-1251",
		"iso-8859-2": "latin2",
	}

	if alias, ok := aliases[normalized]; ok {
		return alias
	}
	return normalized
}

// QuickDetect 快速检测编码（只返回检测到的编码名称）
func QuickDetect(content []byte) (string, error) {
	result, err := defaultDetector.DetectBest(content)
	if err != nil {
		return "", err
	}
	return result.Charset, nil
}
