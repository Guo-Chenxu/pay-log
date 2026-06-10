package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Guo-Chenxu/pay-log/config"
	"github.com/Guo-Chenxu/pay-log/pkg/bizerr"
	"github.com/Guo-Chenxu/pay-log/pkg/logger"
)

// 中间件 ErrorHandlerMiddleware 捕获 panic 错误.
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 保存原始的 ResponseWriter，以便在 panic 时可以直接写入
		originalWriter := c.Writer
		startTime := time.Now()

		defer func() {
			if r := recover(); r != nil {
				// 获取详细的panic信息和堆栈
				stack := string(debug.Stack())
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%+v", r)
				}

				// 读取并恢复请求体
				requestBody, _ := io.ReadAll(c.Request.Body)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
				requestBodyMap := make(map[string]interface{})
				if len(requestBody) > 0 {
					_ = json.Unmarshal(requestBody, &requestBodyMap)
				}

				endTime := time.Now()
				latencyTime := endTime.Sub(startTime)

				c.Header("X-Response-Time", strconv.FormatInt(time.Since(startTime).Milliseconds(), 10))

				logFields := map[string]interface{}{
					"request": map[string]interface{}{
						"request_time": startTime.Format("2006-01-02 15:04:05.000"),
						"client_ip":    c.ClientIP(),
						"token":        c.GetHeader(config.GetAuthConfig().TokenName),
						"method":       c.Request.Method,
						"uri":          c.Request.RequestURI,
						"host":         c.Request.Host,
						"user_agent":   c.Request.UserAgent(),
						"query_params": c.Request.URL.Query(),
						"request_body": requestBodyMap,
						"errors":       c.Errors.ByType(gin.ErrorTypePublic),
					},
					"response": map[string]interface{}{
						"response_time": endTime.Format("2006-01-02 15:04:05.000"),
						"latency_ms":    latencyTime.Milliseconds(),
						"status_code":   c.Writer.Status(),
						"size":          c.Writer.Size(),
						"written":       c.Writer.Written(),
					},
					"error": map[string]interface{}{
						"error":      err,
						"stack":      stack,
						"error_type": "panic",
					},
				}

				// 记录panic日志
				logger.WithFields(logFields).WithContext(c.Request.Context()).Error("Panic Recovered")

				c.Abort()

				// 构造错误响应
				traceID, _ := logger.GetFromContext(c.Request.Context(), logger.TraceIDKey)
				errorResponse := gin.H{
					"code":     bizerr.Fail.Code,
					"message":  bizerr.Fail.Message,
					"trace_id": traceID,
				}

				jsonBytes, _ := json.Marshal(errorResponse)

				c.Writer = originalWriter

				c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
				c.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", len(jsonBytes)))

				_, _ = c.Writer.Write(jsonBytes)
			}
		}()

		c.Next()
	}
}
