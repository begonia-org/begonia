package logger

import (
	"fmt"
	"strings"
	"sync"

	common "github.com/begonia-org/go-sdk/common/api/v1"
	"github.com/begonia-org/go-sdk/logger"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
)

type loggerFormatter struct{}

// var Logger logger.Logger = nil
var onceLog sync.Once

type errHook struct{}

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
	return &LoggerImpl{Entry: l.Entry.WithFields(fields)}
}
func (l *LoggerImpl) Logurs() *logrus.Logger {
	return l.Logger
}
func (l *LoggerImpl) Info(args ...interface{}) {
	l.Entry.Info(args...)
}
func (l *LoggerImpl) Error(err error) {
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
					logger.Error(err)
					return
				}
			}
		}

	}
	l.Entry.Error(err)
}
func (f *loggerFormatter) getFormatterFields(data logrus.Fields) string {
	fields := make([]string, 0)
	entryKeys := []string{"name", "x-uid", "x-request-id", "uri", "method", "remote_addr", "status", "elapsed", "file", "line", "fn"}
	for _, v := range entryKeys {
		if data[v] == nil {
			continue
		}
		fields = append(fields, fmt.Sprintf("%v", data[v]))
	}
	return strings.Join(fields, "|")

}

// Format 实现 logrus.Formatter 接口
func (f *loggerFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 自定义时间格式
	timestamp := entry.Time.Format("2006-01-02 15:04:05")
	// 构建日志字符串
	fieldsStr := ""
	fields := f.getFormatterFields(entry.Data)
	if fields != "" {
		fieldsStr = "|" + fields
	}
	logMessage := fmt.Sprintf("%s|%s|%s%s\n", strings.ToUpper(entry.Level.String()), timestamp, strings.TrimRight(entry.Message, "\n"), fieldsStr)

	return []byte(logMessage), nil
}

func init() {
	onceLog.Do(
		func() {
			_logger := logrus.New()
			_logger.SetFormatter(new(loggerFormatter))
			_logger.SetReportCaller(true)
			Log = &LoggerImpl{_logger.WithFields(logrus.Fields{})}
		},
	)
}
