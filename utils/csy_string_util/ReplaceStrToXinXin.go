package csy_string_util

// ReplaceStrToXinXin 字符串固定长度转 *
func ReplaceStrToXinXin(str string, start, end int) (result string) {
	runes := []rune(str)
	var sfz []rune
	if len(runes) > start {
		sfz = append(sfz, runes[0:start]...)
	}
	if len(runes) > start+end {
		size := len(runes) - start + end
		d := ""
		for i := 0; i < size; i++ {
			d += "*"
		}
		sfz = append(sfz, []rune(d)...)
	}
	if len(str) > start+end {
		sfz = append(sfz, runes[len(runes)-2:len(runes)]...)
	}

	if len(sfz) != 0 {
		return string(sfz)
	} else {
		return str
	}
}
