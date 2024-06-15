package gateway

import (
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/gorilla/websocket"
	c "github.com/smartystreets/goconvey/convey"
)

func TestWebsocketForwarder(t *testing.T) {
	c.Convey("test websocket forwarder", t, func() {
		wk := &websocketForwarder{
			websocket: &websocket.Conn{},
		}
		patch:=gomonkey.ApplyFuncReturn((*websocket.Conn).WriteMessage, fmt.Errorf("write error"))
		defer patch.Reset()
		_,err := wk.Write([]byte("test"))
		c.So(err, c.ShouldNotBeNil)
		c.So(err.Error(), c.ShouldContainSubstring, "write error")
		patch.Reset()

	})
}
