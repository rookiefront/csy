package define_api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BasicContext struct {
	*gin.Context
}

type ResultJSON struct {
	Err  error       `json:"err"`
	Data interface{} `json:"data"`
}

// formatError 统一将 any 类型的 err 转为字符串，安全处理 nil 与 error 类型
func formatError(err any) string {
	if err == nil {
		return ""
	}
	if e, ok := err.(error); ok {
		return e.Error()
	}
	return fmt.Sprintf("%v", err)
}

// getSafeBodyBytes 安全读取 Request.Body 并重新填充，防止 Body 被消耗后导致后续无法再次获取
func getSafeBodyBytes(c *gin.Context) []byte {
	if c.Request == nil || c.Request.Body == nil {
		return nil
	}
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil
	}
	// 还原 Body
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	return bodyBytes
}

func (c *BasicContext) SendJsonToastOk(data ...interface{}) {
	var message interface{} = "ok"
	var outputData interface{}
	if len(data) >= 1 {
		message = data[0]
	}
	if len(data) >= 2 {
		outputData = data[1]
	}
	c.JSON(http.StatusOK, gin.H{
		"toast": true,
		"msg":   message,
		"code":  200,
		"data":  outputData,
		"where": c.GetReqData(),
	})
}

func (c *BasicContext) SendJsonOk(data ...interface{}) {
	var message interface{} = "ok"
	var outputData interface{}
	if len(data) >= 1 {
		outputData = data[0]
	}
	if len(data) >= 2 {
		message = data[1]
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":   message,
		"code":  200,
		"data":  outputData,
		"where": c.GetReqData(),
	})
}

func (c *BasicContext) SendJsonOkWs(data ...interface{}) []byte {
	var outputData interface{}
	if len(data) >= 1 {
		outputData = data[0]
	}

	marshal, _ := json.Marshal(gin.H{
		"msg":   "ok",
		"code":  200,
		"data":  outputData,
		"where": c.GetReqData(),
	})
	return marshal
}

func (c *BasicContext) SendJsonErr(err any) {
	c.SendJsonErrCode(err, 500)
}

func (c *BasicContext) SendJsonErrCode(err any, code any) {
	c.JSON(http.StatusOK, gin.H{
		"msg":   formatError(err),
		"code":  code,
		"data":  nil,
		"where": c.GetReqData(),
	})
}

func (c *BasicContext) SendJsonErrWs(err any) []byte {
	marshal, _ := json.Marshal(gin.H{
		"msg":   formatError(err),
		"code":  500,
		"data":  nil,
		"where": c.GetReqData(),
	})
	return marshal
}

func WrapHandler(handler func(c *BasicContext)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		myCtx := &BasicContext{Context: ctx}
		handler(myCtx)
	}
}

func (c *BasicContext) GetPostFormParams() (map[string]any, error) {
	if c.Request == nil {
		return make(map[string]any), nil
	}

	// 优先解析普通 Form，若失败再尝试 Multipart Form
	if err := c.Request.ParseForm(); err != nil {
		if err := c.Request.ParseMultipartForm(32 << 20); err != nil && !errors.Is(err, http.ErrNotMultipart) {
			return nil, err
		}
	}

	postMap := make(map[string]any, len(c.Request.PostForm))
	for k, v := range c.Request.PostForm {
		if len(v) > 1 {
			postMap[k] = v
		} else if len(v) == 1 {
			postMap[k] = v[0]
		}
	}

	return postMap, nil
}

func (c *BasicContext) GetQueryParams() map[string]any {
	if c.Request == nil || c.Request.URL == nil {
		return make(map[string]any)
	}
	query := c.Request.URL.Query()
	queryMap := make(map[string]any, len(query))
	for k, v := range query {
		if len(v) > 1 {
			queryMap[k] = v
		} else if len(v) == 1 {
			queryMap[k] = v[0]
		}
	}
	return queryMap
}

// GetReqData 获得请求参数 GET POST FormData JSON 合并
func (c *BasicContext) GetReqData() (reqData map[string]any) {
	query := c.GetQueryParams()
	postQuery, err := c.GetPostFormParams()
	if err == nil {
		for m, v := range postQuery {
			query[m] = v
		}
	}

	// 安全读取并解包 JSON Body，避免直接使用 ShouldBindJSON 损坏 Request.Body
	bodyBytes := getSafeBodyBytes(c.Context)
	if len(bodyBytes) > 0 {
		var jsonData map[string]any
		if err := json.Unmarshal(bodyBytes, &jsonData); err == nil {
			for m, v := range jsonData {
				query[m] = v
			}
		}
	}
	return query
}

func (c *BasicContext) GetToken() string {
	return c.GetHeader("X-Token")
}

func (c *BasicContext) GetRequestValue(key string) string {
	if val := c.Query(key); val != "" {
		return val
	}
	if val := c.PostForm(key); val != "" {
		return val
	}

	bodyBytes := getSafeBodyBytes(c.Context)
	if len(bodyBytes) > 0 {
		var bodyMap map[string]interface{}
		if json.Unmarshal(bodyBytes, &bodyMap) == nil {
			if val, ok := bodyMap[key]; ok {
				if strVal, isStr := val.(string); isStr {
					return strVal
				}
				if filterArr, isSlice := val.([]interface{}); isSlice {
					bytesVal, _ := json.Marshal(filterArr)
					return string(bytesVal)
				}
				return fmt.Sprintf("%v", val)
			}
		}
	}
	return ""
}
