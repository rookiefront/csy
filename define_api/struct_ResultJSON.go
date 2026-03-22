package define_api

import (
	"encoding/json"
	"fmt"
)

func NewResultJSON() ResultJSON {
	return ResultJSON{
		Data: nil,
	}
}

func (j ResultJSON) MarshalJSON() ([]byte, error) {
	h := map[string]interface{}{
		"err":  j.Err,
		"data": j.Data,
	}

	if j.Err != nil {
		h["err"] = j.Err.Error()
	}

	return json.Marshal(h)
}

func (c ResultJSON) ToJSON() string {
	marshal, err := json.Marshal(c)
	if err != nil {
		fmt.Println(err)
		return "{}"
	}
	return string(marshal)
}

//func GetResponseBody(resp *http.Response) ([]byte, error) {
//	if resp == nil || resp.Body == nil {
//		return nil, fmt.Errorf("invalid response")
//	}
//
//	// 保存原始读取器以便稍后关闭
//	originalBody := resp.Body
//
//	// 创建一个缓冲区来存储读取的数据
//	var bodyBuffer bytes.Buffer
//
//	// 处理可能的压缩
//	var reader io.Reader = originalBody
//	if resp.Header.Get("Content-Encoding") == "gzip" {
//		gzReader, err := gzip.NewReader(originalBody)
//		if err != nil {
//			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
//		}
//		defer gzReader.Close()
//		reader = gzReader
//	}
//
//	// 使用 TeeReader 同时读取和复制
//	teeReader := io.TeeReader(reader, &bodyBuffer)
//
//	// 先读取一次（这会填充缓冲区）
//	tempBody, err := io.ReadAll(teeReader)
//	if err != nil {
//		return nil, fmt.Errorf("failed to read response body: %w", err)
//	}
//
//	// 获取实际的字节
//	body := bodyBuffer.Bytes()
//
//	// 重要：重置响应体以便后续使用
//	resp.Body = io.NopCloser(bytes.NewReader(body))
//
//	// 更新头部信息
//	if resp.Header.Get("Content-Encoding") == "gzip" {
//		resp.Header.Del("Content-Encoding")
//		resp.Header.Del("Content-Length")
//		resp.ContentLength = int64(len(body))
//	}
//
//	// 关闭原始读取器
//	if originalBody != resp.Body {
//		originalBody.Close()
//	}
//
//	// 验证读取的数据
//	if len(tempBody) != len(body) {
//		return body, fmt.Errorf("data length mismatch: tee=%d, buffer=%d",
//			len(tempBody), len(body))
//	}
//
//	return body, nil
//}
