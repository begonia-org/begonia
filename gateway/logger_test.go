package gateway

import (
	"context"
	"testing"

	c "github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc/metadata"
)

func TestLog(t *testing.T) {
	c.Convey("TestLog", t, func() {

		Log.Info(context.Background(), "info")
		Log.Warn(context.Background(), "warn")
		Log.Infof(context.Background(), "infof")
		Log.Debug(context.Background(), "debug")
		Log.Debugf(context.Background(), "debugf")
		Log.Info(metadata.NewIncomingContext(context.Background(), metadata.Pairs("app", "test")), "infof")
		Log.Logurs().Info("info")
		Log.SetReportCaller(true)

		loggerMid := NewLoggerMiddleware(Log)
		loggerMid.SetPriority(2)
		c.So(loggerMid.Priority(), c.ShouldEqual, 2)
		c.So(loggerMid.Name(), c.ShouldEqual, "logger")

	})
}
