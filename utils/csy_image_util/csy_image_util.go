package csy_image_util

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"regexp"
	"strings"

	"github.com/disintegration/imaging"
)

func Resize(img *image.RGBA, width int) (resizedImg image.Image) {
	if width > 0 {
		bounds := img.Bounds()
		originalWidth := bounds.Dx()
		originalHeight := bounds.Dy()
		newHeight := int(float64(originalHeight) * (float64(width) / float64(originalWidth)))
		resizedImg = imaging.Resize(img, width, newHeight, imaging.Lanczos)
	} else {
		resizedImg = img
	}
	return resizedImg
}

func ToBase64Str(img image.Image) (string, error) {
	// 编码为 PNG
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}
	base64Str := base64.StdEncoding.EncodeToString(buf.Bytes())
	// 添加 Data URL 前缀
	return "data:image/png;base64," + base64Str, nil
}
func ByteToBase64Str(buf []byte) (string, error) {
	base64Str := base64.StdEncoding.EncodeToString(buf)
	// 添加 Data URL 前缀
	return "data:image/png;base64," + base64Str, nil
}

func Base64toImage(base64Str string) (image.Image, error) {
	base64Str = Base64RemoveFirstTag(base64Str)
	// 解码 Base64 字符串
	decoded, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return nil, err
	}

	// 创建图像对象
	img, err := imaging.Decode(bytes.NewReader(decoded))
	if err != nil {
		return nil, err
	}
	return img, nil
}

func base64ReplaceEncode(data []byte) string {
	str := base64.StdEncoding.EncodeToString(data)
	str = strings.Replace(str, "+", "-", -1)
	str = strings.Replace(str, "/", "_", -1)
	str = strings.Replace(str, "=", "", -1)
	return str
}

func base64ReplaceDecode(str string) ([]byte, error) {
	str = strings.Replace(str, "-", "+", -1)
	str = strings.Replace(str, "_", "/", -1)
	for len(str)%4 != 0 {
		str += "="
	}
	return base64.StdEncoding.DecodeString(str)
}

func Base64RemoveFirstTag(base64Str string) string {

	// 使用正则表达式提取 Base64 字符串部分
	re := regexp.MustCompile(`^data:image\/\w+;base64,(.+)`)
	matches := re.FindStringSubmatch(base64Str)

	if len(matches) > 1 {
		base64Str = matches[1]
	}

	return base64Str
}
