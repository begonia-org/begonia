package logger

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
)

type loggerFormatter struct{}

var Logger *logrus.Logger = nil
var onceLog sync.Once

func (f *loggerFormatter) getFormatterFields(data logrus.Fields) string {
	fields := make([]string, 0)
	entryKeys := []string{"name", "x-uid", "x-request-id", "uri", "method", "remote_addr", "status","elapsed"}
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
func LoggerFromContext(ctx context.Context) *logrus.Entry {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {

		return Logger.WithFields(logrus.Fields{})
	}
	uri, _ := runtime.HTTPPathPattern(ctx)
	method, _ := runtime.RPCMethod(ctx)
	remoteAddr := ""
	origin := ""

	if val := md.Get("grpcgateway-origin"); len(val) > 0 {
		origin = val[0]
	}
	if val := md.Get("grpcgateway-referer"); len(val) > 0 {
		remoteAddr = val[0]
	}
	return Logger.WithFields(logrus.Fields{
		"x-request-id": md.Get("x-request-id")[0],
		"remote_addr":  remoteAddr,
		"origin":       origin,
		"uri":          uri,
		"method":       method,
	})
}
func init() {
	onceLog.Do(
		func() {
			Logger = logrus.New()
			Logger.SetFormatter(new(loggerFormatter))
			Logger.SetReportCaller(true)
		},
	)
}
