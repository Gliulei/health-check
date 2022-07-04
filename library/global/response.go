package global

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	ERROR   = 400
	SUCCESS = 0
)

//var Db = g.Cfg("system").GetString("system.Db")

// 数据返回通用JSON数据结构
type JsonResponse struct {
	Code    int         `json:"errcode"` // 错误码((0:成功, 1:失败, >1:错误码))
	Data    interface{} `json:"result"` // 返回数据(业务接口定义具体数据结构)
	Message string      `json:"errmsg"`  // 提示信息
}

func Result(r *gin.Context, code int, data interface{}, message string) {
	r.JSON(http.StatusOK, JsonResponse{
		code,
		data,
		message,
	})
}

func Ok(r *gin.Context) {
	Result(r, SUCCESS, map[string]interface{}{}, "操作成功")
}

func OkWithMessage(r *gin.Context, message string) {
	Result(r, SUCCESS, map[string]interface{}{}, message)
}

func OkWithData(r *gin.Context, data interface{}) {
	Result(r, SUCCESS, data, "操作成功")
}

func OkDetailed(r *gin.Context, data interface{}, message string) {
	Result(r, SUCCESS, data, message)
}

func Fail(r *gin.Context) {
	Result(r, ERROR, map[string]interface{}{}, "操作失败")
}

func FailWithMessage(r *gin.Context, message string) {
	Result(r, ERROR, map[string]interface{}{}, message)
}

func FailWithDetailed(r *gin.Context, code int, data interface{}, message string) {
	Result(r, code, data, message)
}
