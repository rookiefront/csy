package csy_formatcode_util

import "go/format"

func FormatCode(src []byte) ([]byte, error) {
	got, err := format.Source(src)
	if len(src) == 0 {
		return src, nil
	}
	if len(got) == 0 {
		return src, nil
	}

	return got, err
}
