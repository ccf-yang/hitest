package helper

import (
	"compress/gzip"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"hitest/internal/agent/global"

	"hitest/internal/agent/app/dto"
	"hitest/internal/agent/i18n"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func ErrorWithDetail(ctx *gin.Context, code int, msgKey string, err error) {
	res := dto.Response{
		Code:    code,
		Message: "",
	}
	res.Message = i18n.GetMsgWithDetail(msgKey, err.Error())
	ctx.JSON(http.StatusOK, res)
	ctx.Abort()
}

func InternalServer(ctx *gin.Context, err error) {
	ErrorWithDetail(ctx, http.StatusInternalServerError, "ErrInternalServer", err)
}

func BadRequest(ctx *gin.Context, err error) {
	ErrorWithDetail(ctx, http.StatusBadRequest, "ErrInvalidParams", err)
}

func SuccessWithData(ctx *gin.Context, data interface{}) {
	if data == nil {
		data = gin.H{}
	}
	res := dto.Response{
		Code: http.StatusOK,
		Data: data,
	}
	ctx.JSON(http.StatusOK, res)
	ctx.Abort()
}

var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(nil)
	},
}

func SuccessWithDataGzipped(ctx *gin.Context, data interface{}) {
	if !strings.Contains(ctx.GetHeader("Accept-Encoding"), "gzip") {
		SuccessWithData(ctx, data)
		return
	}
	if data == nil {
		data = gin.H{}
	}
	res := dto.Response{
		Code: http.StatusOK,
		Data: data,
	}
	jsonBytes, err := json.Marshal(res)
	if err != nil {
		ErrorWithDetail(ctx, http.StatusInternalServerError, "ErrInternalServer", err)
		return
	}

	ctx.Header("Content-Encoding", "gzip")
	ctx.Header("Content-Type", "application/json; charset=utf-8")
	ctx.Status(http.StatusOK)

	gz := gzipWriterPool.Get().(*gzip.Writer)
	gz.Reset(ctx.Writer)
	_, _ = gz.Write(jsonBytes)
	_ = gz.Close()
	gzipWriterPool.Put(gz)
	ctx.Abort()
}

func Success(ctx *gin.Context) {
	res := dto.Response{
		Code:    http.StatusOK,
		Message: "success",
	}
	ctx.JSON(http.StatusOK, res)
	ctx.Abort()
}

func CheckBindAndValidate(req interface{}, c *gin.Context) error {
	if err := c.ShouldBindJSON(req); err != nil {
		ErrorWithDetail(c, http.StatusBadRequest, "ErrInvalidParams", err)
		return err
	}
	if err := global.VALID.Struct(req); err != nil {
		ErrorWithDetail(c, http.StatusBadRequest, "ErrInvalidParams", err)
		return err
	}
	return nil
}

func CheckBind(req interface{}, c *gin.Context) error {
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorWithDetail(c, http.StatusBadRequest, "ErrInvalidParams", err)
		return err
	}
	return nil
}

func GetParamInt32(paramName string, c *gin.Context) (int32, error) {
	idParam, ok := c.Params.Get(paramName)
	if !ok {
		return 0, errors.New("error id in path")
	}
	intNum, _ := strconv.Atoi(idParam)
	return int32(intNum), nil
}
