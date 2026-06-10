package bizerr

type ReturnCode struct {
	Message string
	Code    int
}

var (
	Success = ReturnCode{"成功", 0}
	Fail    = ReturnCode{"系统繁忙,请稍后再试", -1}

	// 客户端错误 4xx.
	ParamException   = ReturnCode{"参数错误", 400}
	MethodNotAllowed = ReturnCode{"请求方法不支持", 405}
	Unauthorized     = ReturnCode{"未授权", 401}
	Forbidden        = ReturnCode{"您没有相关的权限", 403}
	NotFound         = ReturnCode{"您访问的对象不存在", 404}

	// 服务端通用错误 5xx.
	ServerError   = ReturnCode{"服务器错误", 500}
	DBError       = ReturnCode{"数据库访问错误", 5001}
	ServerBusy    = ReturnCode{"服务器繁忙，请稍后再试", 5002}
)
