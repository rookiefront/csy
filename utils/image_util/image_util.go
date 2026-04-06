package csy_image_util

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"

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
