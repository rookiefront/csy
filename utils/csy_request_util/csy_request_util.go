package csy_request_util

import (
	"resty.dev/v3"
)

// https://resty.dev/
func NewRequest() *resty.Client {
	client := resty.New()
	return client
}

func SendFormData(url string, data map[string]string) (*resty.Response, error) {
	client := resty.New()
	defer client.Close()
	res, err := client.R().
		SetFormData(data).
		Post(url)
	return res, err
}
