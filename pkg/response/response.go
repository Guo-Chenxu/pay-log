package response

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/Guo-Chenxu/pay-log/pkg/bizerr"
	"github.com/Guo-Chenxu/pay-log/pkg/logger"
)

// Response 标准 API 响应格式.
type Response[T any] struct {
	Data    T      `json:"data,omitempty"`
	Message string `json:"message" example:"success"`
	Code    int    `json:"code" example:"200"`
}

// CommonResponse 为了 Swagger 文档的别名.
type CommonResponse = Response[any]

// PaginatedData 分页响应数据格式.
type PaginatedData[T any] struct {
	Items    []T   `json:"items"`
	Page     int   `json:"page" example:"1"`
	PageSize int   `json:"page_size" example:"10"`
	Total    int64 `json:"total,string" example:"100"`
	HasNext  bool  `json:"has_next" example:"true"`
}

func EmptyPageData[T any](page, pageSize int, total int64) *PaginatedData[T] {
	return &PaginatedData[T]{
		HasNext:  false,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		Items:    []T{},
	}
}

// HandleCommonResponse 统一的响应处理（封装了成功响应、业务逻辑错误响应、HTTP错误响应）.
func HandleCommonResponse(ctx *gin.Context, data any, err error) {
	if err != nil {
		var customErr *bizerr.CustomError
		if errors.As(err, &customErr) {
			FailResponse(ctx, customErr.Code, customErr.Message, data)
			return
		}
		// 如果不是自定义错误，则返回通用的500错误
		logger.CtxErrorf(ctx.Request.Context(), "Internal Server Error: %+v", err)
		FailResponse(ctx, bizerr.Fail.Code, bizerr.Fail.Message)
		return
	}
	SuccessResponse(ctx, data)
}

// HandleInvalidParamsResponse 处理参数错误响应.
func HandleInvalidParamsResponse(ctx *gin.Context, errMsg string) {
	if errMsg == "" {
		FailResponse(ctx, bizerr.ParamException.Code, bizerr.ParamException.Message)
		return
	}
	FailResponse(ctx, bizerr.ParamException.Code, bizerr.ParamException.Message+": "+errMsg)
}

// SuccessResponse 成功响应.
func SuccessResponse(c *gin.Context, data any) {
	c.JSON(200, Response[any]{Code: bizerr.Success.Code, Message: bizerr.Success.Message, Data: data})
}

// FailResponse 业务逻辑错误响应.
func FailResponse(c *gin.Context, code int, message string, data ...any) {
	if len(data) > 0 {
		c.JSON(200, Response[any]{Code: code, Message: message, Data: data[0]})
		return
	}
	c.JSON(200, Response[any]{Code: code, Message: message})
}

// UnauthorizedResponse 未授权的错误响应.
func UnauthorizedResponse(c *gin.Context) {
	FailResponse(c, bizerr.Unauthorized.Code, bizerr.Unauthorized.Message)
}
