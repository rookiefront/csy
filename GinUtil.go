package csy

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
)

type GisUtil struct {
}

func NewGisUtil() *GisUtil {
	return &GisUtil{}
}
func (GisUtil) GetBody(res *http.Response) ([]byte, error) {
	var body []byte
	var err error
	// 是否是 gzip 编码
	if res.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(res.Body)
		if err != nil {
			return body, err
		}
		defer gzReader.Close()
		body, err = io.ReadAll(gzReader)
		if err != nil {
			return body, err
		}
		res.Header.Del("Content-Encoding") // 移除 gzip 头
	} else {
		body, err = io.ReadAll(res.Body)
		if err != nil {
			return body, err
		}
	}
	// 回写body数据
	res.Body = io.NopCloser(bytes.NewReader(body))
	return body, err
}
