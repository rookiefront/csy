package csy_string_util

import (
	"regexp"
	"strings"
	"unicode"
)

func StrFirstToUpper(str string) string {
	temp := strings.Split(str, "_")
	var upperStr string
	for y := 0; y < len(temp); y++ {
		vv := []rune(temp[y])
		if y != 0 {
			for i := 0; i < len(vv); i++ {
				if i == 0 {
					vv[i] -= 32
					upperStr += string(vv[i]) // + string(vv[i+1])
				} else {
					upperStr += string(vv[i])
				}
			}
		}
	}
	return temp[0] + upperStr
}

// Capitalize 字符首字母大写
func StrCapitalize(str string) string {
	var upperStr string
	vv := []rune(str) // 后文有介绍
	for i := 0; i < len(vv); i++ {
		if i == 0 {
			if vv[i] >= 97 && vv[i] <= 122 { // 后文有介绍
				vv[i] -= 32 // string的码表相差32位
				upperStr += string(vv[i])
			} else {
				return str
			}
		} else {
			upperStr += string(vv[i])
		}
	}
	return upperStr
}

func StrFirstToLower(str string) string {
	if len(str) == 0 {
		return str
	}
	runStr := []rune(str)
	return strings.ToLower(string(runStr[:1])) + string(runStr[1:])
}

func FieldConvToFrontField(string2 string) string {
	return regexp.MustCompile(`ID$`).ReplaceAllString(StrFirstToLower(string2), "Id")
}

func StrUpperToSplit(str string, splitStr string) string {
	var result string
	for i, r := range str {
		if unicode.IsUpper(r) {
			if i > 0 {
				result += splitStr
			}
			result += string(unicode.ToLower(r))
		} else {
			result += string(r)
		}
	}
	return result

}
