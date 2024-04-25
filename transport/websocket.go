package transport

import (
	"bytes"
	"net/http"

	"github.com/gorilla/websocket"
)

type WebsocketForwarder interface {
	http.ResponseWriter
	Read() ([]byte, error)
	Write([]byte) (int, error)
	Close() error
}

type websocketForwarder struct {
	http.ResponseWriter
	websocket    *websocket.Conn
	responseType int
	stream       StreamClient
}

func NewWebsocketForwarder(w http.ResponseWriter, req *http.Request,
	stream StreamClient, responseType int) (WebsocketForwarder, error) {
	var upgrader = websocket.Upgrader{
		// 允许所有CORS请求
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		return nil, err
	}
	return &websocketForwarder{w, conn, responseType, stream}, nil
}
func (w *websocketForwarder) Flush() {
	// w.ResponseWriter.(http.Flusher).Flush()
}
func (w *websocketForwarder) Close() error {
	_=w.stream.CloseSend()
	return w.websocket.Close()
}
func (w *websocketForwarder) Read() ([]byte, error) {
	_, msg, err := w.websocket.ReadMessage() // 读取消息
	if err != nil {
		return nil, err
	}
	return msg, nil

}
func (w *websocketForwarder) Write(message []byte) (int, error) {
	if len(message) == 0 || bytes.Equal(message, []byte("\n")) {
		return 0, nil
	}
	err := w.websocket.WriteMessage(w.responseType, message)
	if err != nil {
		return 0, err
	}
	return len(message), err
}
