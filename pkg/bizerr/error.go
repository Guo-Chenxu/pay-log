package bizerr

import "fmt"

type CustomError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// 实现 error 接口.
var _ error = (*CustomError)(nil)

func (e *CustomError) Error() string {
	return fmt.Sprintf("Code: %d, Message: %s", e.Code, e.Message)
}

// 创建一个新的自定义错误.
func NewCustomError(returnCode ReturnCode) *CustomError {
	return &CustomError{
		Code:    returnCode.Code,
		Message: returnCode.Message,
	}
}

// 创建一个新的自定义错误，带有额外的消息.
func NewCustomErrorWithExtra(code int, extraMessage string) *CustomError {
	return &CustomError{
		Code:    code,
		Message: extraMessage,
	}
}
