package logger

import "context"

// 便利方法.
func Debug(args ...interface{}) {
	l.Debug(args...)
}

func CtxDebug(ctx context.Context, args ...interface{}) {
	l.WithContext(ctx).Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	l.Debugf(format, args...)
}

func CtxDebugf(ctx context.Context, format string, args ...interface{}) {
	l.WithContext(ctx).Debugf(format, args...)
}

func Info(args ...interface{}) {
	l.Info(args...)
}

func CtxInfo(ctx context.Context, args ...interface{}) {
	l.WithContext(ctx).Info(args...)
}

func Infof(format string, args ...interface{}) {
	l.Infof(format, args...)
}

func CtxInfof(ctx context.Context, format string, args ...interface{}) {
	l.WithContext(ctx).Infof(format, args...)
}

func Warn(args ...interface{}) {
	l.Warn(args...)
}

func CtxWarn(ctx context.Context, args ...interface{}) {
	l.WithContext(ctx).Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	l.Warnf(format, args...)
}

func CtxWarnf(ctx context.Context, format string, args ...interface{}) {
	l.WithContext(ctx).Warnf(format, args...)
}

func Error(args ...interface{}) {
	l.Error(args...)
}

func CtxError(ctx context.Context, args ...interface{}) {
	l.WithContext(ctx).Error(args...)
}

func Errorf(format string, args ...interface{}) {
	l.Errorf(format, args...)
}

func CtxErrorf(ctx context.Context, format string, args ...interface{}) {
	l.WithContext(ctx).Errorf(format, args...)
}

func Fatal(args ...interface{}) {
	l.Fatal(args...)
}

func CtxFatal(ctx context.Context, args ...interface{}) {
	l.WithContext(ctx).Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	l.Fatalf(format, args...)
}

func CtxFatalf(ctx context.Context, format string, args ...interface{}) {
	l.WithContext(ctx).Fatalf(format, args...)
}
