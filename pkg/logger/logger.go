package logger

import (
	"context"
	"io"
	"os"
	"path"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	TraceIDKey     = "trace_id"
	UserIDKey      = "user_id"
	maxLogFileSize = 100 * 1024 * 1024 // 100MB
)

type LoggerConfig struct {
	Level string `yaml:"level"`
	Path  string `yaml:"path"`
}

var l *logrus.Logger

// RotatingFileWriter 是一个支持文件大小轮转的 Writer.
type RotatingFileWriter struct {
	file        *os.File
	logPath     string
	symlinkPath string
	currentSize int64
	maxSize     int64
	mu          sync.Mutex
}

func (w *RotatingFileWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 检查是否需要轮转
	if w.currentSize+int64(len(p)) > w.maxSize {
		if err = w.rotate(); err != nil {
			return 0, err
		}
	}

	n, err = w.file.Write(p)
	if err == nil {
		w.currentSize += int64(n)
	}
	return n, err
}

func (w *RotatingFileWriter) rotate() error {
	// 关闭当前文件
	if w.file != nil {
		w.file.Close()
	}

	// 创建新文件
	fileName := path.Join(w.logPath, time.Now().Format("2006-01-02-15-04-05")+".log")
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		return err
	}

	w.file = file
	w.currentSize = 0

	// 更新软链接
	os.Remove(w.symlinkPath)
	relativeFileName := path.Base(fileName)
	if err := os.Symlink(relativeFileName, w.symlinkPath); err != nil {
		l.Warnf("创建软链接失败: %+v", err)
	}

	return nil
}

func (w *RotatingFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

func Init(config LoggerConfig) {
	l = logrus.New()

	// 设置日志格式
	l.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000",
		// PrettyPrint:     true,
	})

	// 创建日志目录
	if err := os.MkdirAll(config.Path, 0o777); err != nil {
		l.Fatal(err)
	}

	// 设置日志文件
	fileName := path.Join(config.Path, time.Now().Format("2006-01-02-15-04-05")+".log")
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		l.Fatal(err)
	}

	// 获取文件当前大小
	fileInfo, err := file.Stat()
	if err != nil {
		l.Fatal(err)
	}

	// 创建 system.log 软链接
	symlinkPath := path.Join(config.Path, "system.log")
	// 删除旧的软链接（如果存在）
	os.Remove(symlinkPath)
	// 创建新的软链接，使用相对路径
	relativeFileName := path.Base(fileName)
	if err = os.Symlink(relativeFileName, symlinkPath); err != nil {
		l.Warnf("创建软链接失败: %+v", err)
	}

	// 创建轮转文件写入器
	rotatingWriter := &RotatingFileWriter{
		file:        file,
		currentSize: fileInfo.Size(),
		maxSize:     maxLogFileSize,
		logPath:     config.Path,
		symlinkPath: symlinkPath,
	}

	// 同时输出到文件和控制台
	l.SetOutput(rotatingWriter)
	// 先添加上下文字段 hook，再添加控制台输出 hook
	l.AddHook(&ContextFieldsHook{})
	l.AddHook(&ConsoleHook{})

	// 设置日志级别
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		panic(err)
	}
	l.SetLevel(level)
}

func Output() io.Writer {
	return l.Out
}

func WithFields(fields logrus.Fields) *logrus.Entry {
	return l.WithFields(fields)
}

// 用于在 context 中存储日志字段的 key.
type contextKey struct{}

var contextFieldsKey = &contextKey{}

// AddToContext 向 context 中追加一个日志字段（不可变，返回新的 context）.
func AddToContext(ctx context.Context, key string, value any) context.Context {
	var newFields logrus.Fields
	if existing, ok := ctx.Value(contextFieldsKey).(logrus.Fields); ok {
		newFields = make(logrus.Fields, len(existing)+1)
		for k, v := range existing {
			newFields[k] = v
		}
	} else {
		newFields = make(logrus.Fields, 1)
	}
	newFields[key] = value
	return context.WithValue(ctx, contextFieldsKey, newFields)
}

// GetFromContext 从 context 中获取一个日志字段.
func GetFromContext(ctx context.Context, key string) (any, bool) {
	if fields, ok := ctx.Value(contextFieldsKey).(logrus.Fields); ok {
		if v, ok := fields[key]; ok {
			return v, true
		}
	}
	return nil, false
}

// ContextFieldsHook 在写日志前把 context 中的字段合并进 entry.Data.
type ContextFieldsHook struct{}

func (h *ContextFieldsHook) Levels() []logrus.Level { return logrus.AllLevels }

func (h *ContextFieldsHook) Fire(entry *logrus.Entry) error {
	if entry.Context == nil {
		return nil
	}
	if v := entry.Context.Value(contextFieldsKey); v != nil {
		if fields, ok := v.(logrus.Fields); ok {
			for k, val := range fields {
				if _, exists := entry.Data[k]; !exists {
					entry.Data[k] = val
				}
			}
		}
	}
	return nil
}

// ConsoleHook 用于同时将日志输出到控制台.
type ConsoleHook struct{}

func (hook *ConsoleHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err == nil {
		os.Stdout.Write([]byte(line))
	}
	return err
}

func (hook *ConsoleHook) Levels() []logrus.Level {
	return logrus.AllLevels
}
