package response

import (
	"github.com/gin-gonic/gin"
)

type Resp struct {
	Code uint   `json:"code"`
	Data any    `json:"data,omitempty"`
	Msg  string `json:"msg,omitempty"`
}

func Success(c *gin.Context) {
	SuccessWithData(c, nil)
}

func SuccessWithData(c *gin.Context, data any) {
	c.JSON(200, Resp{
		Code: 0,
		Data: data,
		Msg:  "success",
	})
}

func Error(c *gin.Context, code uint) {
	ErrorWithData(c, code, nil)
}

func ErrorWithData(c *gin.Context, code uint, data any) {
	c.JSON(200, Resp{
		Code: code,
		Data: data,
		Msg:  MsgMap[code],
	})
}

func ErrorWithHttpCode(c *gin.Context, httpCode int) {
	c.JSON(httpCode, Resp{
		Code: uint(httpCode),
		Data: nil,
		Msg:  MsgMap[uint(httpCode)],
	})
}

func ErrorWithHttpCodeAndCode(c *gin.Context, httpCode int, code uint) {
	c.JSON(httpCode, Resp{
		Code: code,
		Data: nil,
		Msg:  MsgMap[code],
	})
}
