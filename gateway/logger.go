package gateway

import (
	"context"
	"encoding/json"
	"runtime"
	"strings"
	"sync"

	common "github.com/begonia-org/go-sdk/common/api/v1"
	"github.com/begonia-org/go-sdk/logger"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
)

type loggerFormatter struct{}
type loggerHook struct{}

var onceLog sync.Once

type LoggerImpl struct {
	*logrus.Entry
}

var Log logger.Logger

func (l *LoggerImpl) WithField(key string, value interface{}) logger.Logger {
	return &LoggerImpl{Entry: l.Entry.WithField(key, value)}
}
func (l *LoggerImpl) SetReportCaller(reportCaller bool) {
	l.Logger.SetReportCaller(reportCaller)
}
func (l *LoggerImpl) WithFields(fields logrus.Fields) logger.Logger {
	imp := &LoggerImpl{Entry: l.Entry.WithFields(fields)}
	return imp
}
func (l *LoggerImpl) WithContext(ctx context.Context) logger.Logger {

	return &LoggerImpl{Entry: l.Entry.WithContext(ctx)}
}
func (l *LoggerImpl) Logurs() *logrus.Logger {
	return l.Logger
}
func (l *LoggerImpl) getFieldsFromContext(ctx context.Context) logrus.Fields {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return logrus.Fields{}

	}
	reqId := md.Get(XRequestID)[0]
	remoteAddr := ""
	if len(md.Get("x-forwarded-for")) > 0 {
		remoteAddr = md.Get("x-forwarded-for")[0]
	}
	if remoteAddr == "" && len(md.Get(XRemoteAddr)) > 0 {
		remoteAddr = md.Get(XRemoteAddr)[0]
	}
	method := ""
	uri := ""
	xuid := ""
	xAccessKey := ""
	if len(md.Get(XHttpMethod)) > 0 {
		method = md.Get(XHttpMethod)[0]
	}
	if len(md.Get(XHttpURI)) > 0 {
		uri = md.Get(XHttpURI)[0]
	}
	if len(md.Get(XAccessKey)) > 0 {
		xAccessKey = md.Get(XAccessKey)[0]
	}

	fields := logrus.Fields{
		XRequestID:  reqId,
		XHttpURI:    uri,
		XHttpMethod: method,
		XRemoteAddr: remoteAddr,
		XUID:        xuid,
		XAccessKey:  xAccessKey,
	}
	return fields
}

func (l *LoggerImpl) Info(ctx context.Context, args ...interface{}) {
	l.Entry.WithFields(l.getFieldsFromContext(ctx)).Info(args...)
}

func (l *LoggerImpl) Infof(ctx context.Context, format string, args ...interface{}) {

	l.Entry.WithFields(l.getFieldsFromContext(ctx)).Infof(format, args...)
}
func (l *LoggerImpl) Errorf(ctx context.Context, format string, args ...interface{}) {
	l.Entry.WithFields(l.getFieldsFromContext(ctx)).Errorf(format, args...)
}
func (l *LoggerImpl) Warnf(ctx context.Context, format string, args ...interface{}) {
	l.Entry.WithFields(l.getFieldsFromContext(ctx)).Warnf(format, args...)

}
func (l *LoggerImpl) Debugf(ctx context.Context, format string, args ...interface{}) {
	l.Entry.WithFields(l.getFieldsFromContext(ctx)).Debugf(format, args...)

}
func (l *LoggerImpl) Warn(ctx context.Context, args ...interface{}) {
	l.Entry.WithFields(l.getFieldsFromContext(ctx)).Warn(args...)

}
func (l *LoggerImpl) Debug(ctx context.Context, args ...interface{}) {
	l.Entry.WithFields(l.getFieldsFromContext(ctx)).Debug(args...)
}

func (l *LoggerImpl) errorHandle(ctx context.Context, err error) *logrus.Entry {
	if st, ok := status.FromError(err); ok {
		details := st.Details()
		for _, detail := range details {
			if anyType, ok := detail.(*anypb.Any); ok {
				var errDetail common.Errors
				if err := anyType.UnmarshalTo(&errDetail); err == nil {
					rspCode := float64(errDetail.Code)
					logger := l.Entry.WithFields(logrus.Fields{
						"status": int(rspCode),
						"file":   errDetail.File,
						"line":   errDetail.Line,
						"fn":     errDetail.Fn,
					})
					return logger.WithFields(l.getFieldsFromContext(ctx))

				}
			}
		}

	}
	return l.Entry.WithFields(l.getFieldsFromContext(ctx))
}
func (l *LoggerImpl) Error(ctx context.Context, err error) {
	l.errorHandle(ctx, err).Error(err)
}
func (f *loggerFormatter) getFormatterFields(data logrus.Fields) string {

	bData, _ := json.Marshal(data)
	return string(bData) + "\n"

}

func getPreviousFrame(pc uintptr) (runtime.Frame, bool) {
	pcs := make([]uintptr, 10)
	n := runtime.Callers(6, pcs) // skip runtime.Callers and getPreviousFrame
	frames := runtime.CallersFrames(pcs[:n])

	// Iterate through frames to find the one matching the given PC
	var previousFrame runtime.Frame
	for {
		frame, more := frames.Next()
		if frame.PC == pc {
			// Found the frame, now get the next one
			if more {
				previousFrame, more = frames.Next()
				return previousFrame, more
			}
			break
		}
		if !more {
			break
		}
	}
	return runtime.Frame{}, false
}

// Format 实现 logrus.Formatter 接口
func (f *loggerFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	entry.Data["message"] = entry.Message
	fields := f.getFormatterFields(entry.Data)

	return []byte(fields), nil
}
func (hook *loggerHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *loggerHook) Fire(entry *logrus.Entry) error {
	// 自定义时间格式
	timestamp := entry.Time.Format("2006-01-02 15:04:05")
	entry.Data["timestamp"] = timestamp
	entry.Data["level"] = strings.ToUpper(entry.Level.String())

	if entry.HasCaller() {
		if v := entry.Data["file"]; v == nil {
			caller, ok := getPreviousFrame(entry.Caller.PC)
			if ok {
				entry.Data["file"] = caller.File
				entry.Data["line"] = caller.Line
				entry.Data["fn"] = caller.Func.Name()
				fn := caller.Func.Name()
				if strings.Contains(fn, ".") {
					fn = fn[strings.LastIndex(fn, ".")+1:]
				}
				entry.Data["fn"] = fn
			}
		}
	}

	return nil
}
func init() {
	onceLog.Do(
		func() {
			_logger := logrus.New()
			_logger.SetFormatter(&loggerFormatter{})
			_logger.SetReportCaller(true)
			_logger.AddHook(&loggerHook{})
			Log = &LoggerImpl{_logger.WithFields(logrus.Fields{})}
		},
	)
}
