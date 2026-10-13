package helper

import (
	"fmt"
	"net/http"

	"hitest/internal/core/app/dto"
	"hitest/internal/core/cmd/server/res"
	"hitest/internal/core/i18n"

	"github.com/gin-gonic/gin"
)

func ErrorWithDetail(ctx *gin.Context, code int, msgKey string, err error) {
	res := dto.Response{
		Code:    code,
		Message: "",
	}
	if msgKey == "ErrCaptchaCode" || msgKey == "ErrAuth" || msgKey == "ErrLoginLocked" {
		res.Code = 401
		res.Message = msgKey
	}
	if msgKey == "ErrRBAC" {
		res.Code = 412
		if err != nil {
			res.Message = err.Error()
			ctx.JSON(http.StatusOK, res)
			ctx.Abort()
			return
		}
	}
	res.Message = i18n.GetMsgWithMap(msgKey, map[string]interface{}{"detail": err})
	ctx.JSON(http.StatusOK, res)
	ctx.Abort()
}

func InternalServer(ctx *gin.Context, err error) {
	ErrorWithDetail(ctx, http.StatusInternalServerError, "ErrInternalServer", err)
}

func ErrWithHtml(ctx *gin.Context, code int, scope string) {
	if code == 444 {
		ctx.String(444, "")
		ctx.Abort()
		return
	}
	file := fmt.Sprintf("html/%d.html", code)
	if code == 200 && scope != "" {
		file = fmt.Sprintf("html/200_%s.html", scope)
	}
	data, err := res.ErrorMsg.ReadFile(file)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "Internal Server Error")
		ctx.Abort()
		return
	}
	ctx.Data(code, "text/html; charset=utf-8", data)
	ctx.Abort()
}
