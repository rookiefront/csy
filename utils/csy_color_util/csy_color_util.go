package csy_color_util

import "fmt"

func RgbaToHex(r, g, b uint8, a float64) string {
	// 将透明度转换为 0 到 255 的整数
	alpha := uint8(a * 255)

	// 格式化为 RGBA 的 16 进制字符串
	return fmt.Sprintf("#%02X%02X%02X%02X", r, g, b, alpha)
}
