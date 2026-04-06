package csy_charset_util

import (
	"bytes"
	"io"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
)

// ---------- GBK 系列 ----------

// GbkToUtf8 将 GBK 编码的字节切片转换为 UTF-8 编码的字节切片
func GbkToUtf8(s []byte) ([]byte, error) {
	return decode(s, simplifiedchinese.GBK.NewDecoder())
}

// Utf8ToGbk 将 UTF-8 编码的字节切片转换为 GBK 编码的字节切片
func Utf8ToGbk(s []byte) ([]byte, error) {
	return encode(s, simplifiedchinese.GBK.NewEncoder())
}

// GbkToUtf8String 将 GBK 编码的字节切片转换为 UTF-8 字符串
func GbkToUtf8String(s []byte) (string, error) {
	b, err := GbkToUtf8(s)
	return string(b), err
}

// Utf8ToGbkString 将 UTF-8 字符串转换为 GBK 编码的字节切片
func Utf8ToGbkString(s string) ([]byte, error) {
	return Utf8ToGbk([]byte(s))
}

// ---------- HZ-GB2312 支持 ----------

// HZGB2312ToUtf8 将 HZ-GB2312 编码的字节切片转换为 UTF-8
func HZGB2312ToUtf8(s []byte) ([]byte, error) {
	return decode(s, simplifiedchinese.HZGB2312.NewDecoder())
}

// Utf8ToHZGB2312 将 UTF-8 字节切片转换为 HZ-GB2312 编码
func Utf8ToHZGB2312(s []byte) ([]byte, error) {
	return encode(s, simplifiedchinese.HZGB2312.NewEncoder())
}

// ---------- Big5 (繁体) 支持 ----------
// 注：需要先执行 go get golang.org/x/text/encoding/traditionalchinese

// Big5ToUtf8 将 Big5 编码的字节切片转换为 UTF-8
func Big5ToUtf8(s []byte) ([]byte, error) {
	return decode(s, traditionalchinese.Big5.NewDecoder())
}

// Utf8ToBig5 将 UTF-8 字节切片转换为 Big5 编码
func Utf8ToBig5(s []byte) ([]byte, error) {
	return encode(s, traditionalchinese.Big5.NewEncoder())
}

// ---------- 通用转换函数 (内部使用) ----------

// decode 将编码后的字节切片通过指定的解码器转换为 UTF-8
func decode(s []byte, decoder transform.Transformer) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), decoder)
	return io.ReadAll(reader)
}

// encode 将 UTF-8 字节切片通过指定的编码器转换为目标编码
func encode(s []byte, encoder transform.Transformer) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), encoder)
	return io.ReadAll(reader)
}

// ---------- 流式转换：适用于大文件或网络流 ----------

// NewGbkToUtf8Reader 创建一个将 GBK 流转换为 UTF-8 流的 Reader
func NewGbkToUtf8Reader(r io.Reader) io.Reader {
	return transform.NewReader(r, simplifiedchinese.GBK.NewDecoder())
}

// NewUtf8ToGbkReader 创建一个将 UTF-8 流转换为 GBK 流的 Reader
func NewUtf8ToGbkReader(r io.Reader) io.Reader {
	return transform.NewReader(r, simplifiedchinese.GBK.NewEncoder())
}

// NewGbkToUtf8Writer 创建一个将写入的数据从 GBK 转换为 UTF-8 的 Writer
func NewGbkToUtf8Writer(w io.Writer) io.Writer {
	return transform.NewWriter(w, simplifiedchinese.GBK.NewDecoder())
}

// NewUtf8ToGbkWriter 创建一个将写入的数据从 UTF-8 转换为 GBK 的 Writer
func NewUtf8ToGbkWriter(w io.Writer) io.Writer {
	return transform.NewWriter(w, simplifiedchinese.GBK.NewEncoder())
}

// ---------- 辅助工具 ----------

// MustConvert 转换函数，如果出错则 panic（适用于初始化或测试场景）
func MustConvert(data []byte, convertFunc func([]byte) ([]byte, error)) []byte {
	result, err := convertFunc(data)
	if err != nil {
		panic(err)
	}
	return result
}

// ConvertOrRaw 尝试转换，失败时返回原始数据
func ConvertOrRaw(data []byte, convertFunc func([]byte) ([]byte, error)) []byte {
	result, err := convertFunc(data)
	if err != nil {
		return data
	}
	return result
}
