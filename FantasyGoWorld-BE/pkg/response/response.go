package response

import (
	"fantasy-go-world-be/pkg/e"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
	Code e.Code      `json:"code"`
}

// JSON sends a custom response with HTTP code, business code, and data
func JSON(c *gin.Context, httpCode int, errCode e.Code, data interface{}) {
	if data == nil {
		data = gin.H{}
	}
	c.JSON(httpCode, Response{
		Code: errCode,
		Msg:  e.GetMsg(errCode),
		Data: data,
	})
}

// Success sends a standard 200 OK success response
func Success(c *gin.Context, data interface{}) {
	JSON(c, http.StatusOK, e.SUCCESS, data)
}

// Error sends an error response with custom HTTP and business code
func Error(c *gin.Context, httpCode int, errCode e.Code, data interface{}) {
	JSON(c, httpCode, errCode, data)
}
