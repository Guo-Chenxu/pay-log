package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Guo-Chenxu/pay-log/config"
	"github.com/Guo-Chenxu/pay-log/pkg/logger"
	"github.com/Guo-Chenxu/pay-log/pkg/path"
)

var skipLogPaths = map[string]struct{}{
	"/swagger/*any":     {},
	"/healthz":          {},
	"/api/v1/healthz":   {},
	"/.well-known/*any": {},
}

// 自定义 ResponseWriter 以捕获响应体.
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)    // 捕获响应数据
	return len(b), nil // 先不写入真实响应流
}

// 中间件 LoggerMiddleware 记录请求日志.
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := insertTraceIDToContext(c)
		startTime := time.Now()

		// 读取并恢复请求体
		requestBody, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		requestBodyMap := make(map[string]interface{})
		if len(requestBody) > 0 {
			_ = json.Unmarshal(requestBody, &requestBodyMap)
		}

		// 替换 ResponseWriter 为自定义实现
		// 响应内容会被Write方法写到responseBodyWriter的body字段也就是这个buf
		var buf bytes.Buffer
		writer := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &buf,
		}
		c.Writer = writer

		c.Header("X-Trace-ID", traceID)

		c.Next()

		if c.IsAborted() {
			return
		}

		c.Header("X-Response-Time", strconv.FormatInt(time.Since(startTime).Milliseconds(), 10))

		// 跳过 不打印日志
		if path.Contains(skipLogPaths, c.FullPath()) {
			_, _ = writer.ResponseWriter.Write(buf.Bytes())
			return
		}

		endTime := time.Now()
		latencyTime := endTime.Sub(startTime)

		// 解析响应体中的 JSON 数据, 提取code并注入traceID
		var businessCode int
		var responseBody map[string]interface{}
		finalBody := buf.Bytes()

		if err := json.Unmarshal(buf.Bytes(), &responseBody); err == nil {
			if code, ok := responseBody["code"].(float64); ok {
				businessCode = int(code)
			}
			// 向响应体中注入 traceID
			responseBody["trace_id"] = traceID
			// 重新序列化响应体
			if modifiedBody, err := json.Marshal(responseBody); err == nil {
				finalBody = modifiedBody
				// 更新 Content-Length
				c.Writer.Header().Set("Content-Length", strconv.Itoa(len(modifiedBody)))
			}
		}

		// 将最终的响应体写入真实的响应流
		_, _ = writer.ResponseWriter.Write(finalBody)

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
				"response_time":    endTime.Format("2006-01-02 15:04:05.000"),
				"latency_ms":       latencyTime.Milliseconds(),
				"status_code":      c.Writer.Status(),
				"response_size":    len(finalBody),
				"response_headers": c.Writer.Header(),
				"business_code":    businessCode,
				"response_body":    responseBody,
			},
		}

		logger.WithFields(logFields).WithContext(c.Request.Context()).Info("Access log")
	}
}

func insertTraceIDToContext(c *gin.Context) string {
	traceID := c.GetHeader("X-Trace-ID")
	if traceID == "" {
		traceID = strings.ReplaceAll(uuid.New().String(), "-", "")
		c.Request.Header.Set("X-Trace-ID", traceID)
	}
	c.Set("X-Trace-ID", traceID)
	c.Request = c.Request.WithContext(logger.AddToContext(c.Request.Context(), logger.TraceIDKey, traceID))
	return traceID
}
