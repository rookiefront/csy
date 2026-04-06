package req

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type Request struct {
}

func NewRequest() *Request {
	return &Request{}
}

func (*Request) PostByJson(reqUrl string, jsonData any, headers map[string]string) (responBytes []byte, err error) {
	jsonDataMar, err := json.Marshal(jsonData)
	if err != nil {
		return responBytes, err
	}
	req, err := http.NewRequest("POST", reqUrl, bytes.NewBuffer(jsonDataMar))
	if err != nil {
		return responBytes, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return responBytes, err
	}
	defer resp.Body.Close()
	responBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		return responBytes, err
	}
	return responBytes, err
}
