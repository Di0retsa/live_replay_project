package retcode

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

// ErrorCodeGetter 提取错误码的接口
type ErrorCodeGetter interface {
	GetCode() int
}

// Error 通用错误结构体
type Error struct {
	ErrCode int
	ErrMsg  string
}

func (e *Error) Error() string {
	return e.ErrMsg
}

func (e *Error) GetCode() int {
	return e.ErrCode
}

func NewError(code int, msg string) *Error {
	return &Error{ErrCode: code, ErrMsg: msg}
}

// OK 渲染成功相应
func OK(ctx *gin.Context, result interface{}) {
	RenderReply(ctx, result)
}

func Fatal(ctx *gin.Context, e error, msg string) {
	code := GetErrCode(e)
	fmt.Printf("%s, %+v\n", msg, e)
	if msg == "" {
		msg = e.Error()
	}
	RenderErrMsg(ctx, code, msg)
}

func RenderReply(ctx *gin.Context, data interface{}) {
	render(ctx, http.StatusOK, data, nil)
}

func RenderErrMsg(ctx *gin.Context, code int, msg string) {
	render(ctx, code, nil, errors.New(msg))
}

func render(ctx *gin.Context, code int, data interface{}, err error) {
	var msg string
	if err != nil {
		msg = err.Error()
	} else {
		msg = "操作成功！"
	}
	r := gin.H{
		"code": code,
		"msg":  msg,
		"data": data,
	}
	ctx.Set("return_code", code)
	ctx.JSON(code, r)
}

func GetErrCode(err error) int {
	if errGetter, ok := err.(ErrorCodeGetter); ok {
		return errGetter.GetCode()
	}
	return http.StatusInternalServerError
}
