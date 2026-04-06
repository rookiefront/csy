package define_api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/front-ck996/csy/utils/csy_assert_util"
	"github.com/gin-gonic/gin"
)

type BasicContext struct {
	*gin.Context
}

type ResultJSON struct {
	Err  error       `json:"err"`
	Data interface{} `json:"data"`
}

func (c *BasicContext) SendJsonToastOk(data ...interface{}) {
	var message, outputData interface{}
	if len(data) >= 1 {
		message = data[0]
	}
	if len(data) >= 2 {
		outputData = data[1]
	}
	c.JSON(200, gin.H{
		"toast": true,
		"msg":   message,
		"code":  200,
		"data":  outputData,
		"where": c.GetReqData(),
	})
}

func (c *BasicContext) SendJsonOk(data ...interface{}) {
	var message interface{}
	if len(data) >= 1 {
		message = data[0]
	}
	c.JSON(200, gin.H{
		"msg":   "ok",
		"code":  200,
		"data":  message,
		"where": c.GetReqData(),
	})
}

func (c *BasicContext) SendJsonOkWs(data ...interface{}) []byte {
	var message interface{}
	if len(data) >= 1 {
		message = data[0]
	}

	marshal, _ := json.Marshal(gin.H{
		"msg":   "ok",
		"code":  200,
		"data":  message,
		"where": c.GetReqData(),
	})
	return marshal
}

func (c *BasicContext) SendJsonErr(err any) {
	if csy_assert_util.IsError(err) && err != nil {
		err = err.(error).Error()
	}
	c.JSON(200, gin.H{
		"msg":   err,
		"code":  500,
		"data":  nil,
		"where": c.GetReqData(),
	})
}

func (c *BasicContext) SendJsonErrWs(err any) []byte {
	if csy_assert_util.IsError(err) && err != nil {
		err = err.(error).Error()
	}
	marshal, _ := json.Marshal(gin.H{
		"msg":   err,
		"code":  500,
		"data":  nil,
		"where": c.GetReqData(),
	})
	return marshal
}

func (c *BasicContext) SendJsonErrCode(err any, code any) {
	if csy_assert_util.IsError(err) && err != nil {
		err = err.(error).Error()
	}
	c.JSON(200, gin.H{
		"msg":   err,
		"code":  code,
		"data":  nil,
		"where": c.GetReqData(),
	})
}

func WrapHandler(handler func(c *BasicContext)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		myCtx := &BasicContext{Context: ctx}
		handler(myCtx)
	}
}

func (c *BasicContext) GetPostFormParams() (map[string]any, error) {
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		if !errors.Is(err, http.ErrNotMultipart) {
			return nil, err
		}
	}
	var postMap = make(map[string]any, len(c.Request.PostForm))
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
	query := c.Request.URL.Query()
	var queryMap = make(map[string]any, len(query))
	for k := range query {
		queryMap[k] = c.Query(k)
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
	var jsonData map[string]any
	c.ShouldBindJSON(&jsonData)
	for m, v := range jsonData {
		query[m] = v
	}
	return query
}

//// GetRequestHeaders 获取所有的请求头
//func (c *BasicContext) GetRequestHeaders() (reqData map[string]any) {
//	// 记录 headers
//	headers := c.Request.Header
//
//	//query := c.GetQueryParams()
//	//postQuery, err := c.GetPostFormParams()
//	//if err == nil {
//	//	for m, v := range postQuery {
//	//		query[m] = v
//	//	}
//	//}
//	//var jsonData map[string]any
//	//c.ShouldBindJSON(&jsonData)
//	//for m, v := range jsonData {
//	//	query[m] = v
//	//}
//	//return query
//}

func (c *BasicContext) GetToken() string {
	return c.GetHeader("X-Token")
}
