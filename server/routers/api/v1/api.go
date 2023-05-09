package v1

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"strconv"
)

var (
	OKCode      = 200
	AuthCode    = 3000
	RequestCode = 3002
	ExecuteCode = 3003

	CodeMessages = map[int]string{
		OKCode:      "执行成功",
		AuthCode:    "认证失败",
		RequestCode: "请求参数错误",
		ExecuteCode: "执行错误",
	}
)

type Response struct {
	Code   int         `json:"code"`
	Msg    string      `json:"message"`
	Detail string      `json:"detail"`
	Data   interface{} `json:"data"`
}

func NewResponse(code int, data interface{}) *Response {
	res := &Response{
		Code: code,
	}
	res.Msg, _ = CodeMessages[code]
	if code == OKCode {
		res.Data = data
	} else if err, ok := data.(error); ok {
		res.Detail = err.Error()
		log.Error("api err ", res.Detail)
	} else {
		res.Detail, _ = data.(string)
		log.Error("api err ", res.Detail)
	}
	return res
}

type List struct {
	Total     int64       `json:"total"`
	PageTotal int64       `json:"page_total"`
	Items     interface{} `json:"items"`
}

func page(c *gin.Context) (pageNum int64, pageSize int64) {
	pageNum = 1
	pageSize = 10
	if val, err := strconv.ParseInt(c.DefaultQuery("page_num", "1"), 10, 64); err != nil {

	} else if val != 0 {
		pageNum = val
	}
	if val, err := strconv.ParseInt(c.DefaultQuery("page_size", "10"), 10, 64); err != nil {

	} else if val != 0 {
		pageSize = val
	}
	return
}
