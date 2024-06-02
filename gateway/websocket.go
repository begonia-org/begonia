package gateway

import (
	"bytes"
	"io"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type WebsocketForwarder interface {
	http.ResponseWriter
	// Read() ([]byte, error)
	Write([]byte) (int, error)
	Close() error
	CloseConn() error
	NextReader() (io.Reader, error)
}

type websocketForwarder struct {
	http.ResponseWriter
	websocket    *websocket.Conn
	responseType int
	mux          sync.Mutex
}

func NewWebsocketForwarder(w http.ResponseWriter, req *http.Request, responseType int) (WebsocketForwarder, error) {
	var upgrader = websocket.Upgrader{
		// 允许所有CORS请求
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		return nil, err
	}
	return &websocketForwarder{w, conn, responseType, sync.Mutex{}}, nil
}
func (w *websocketForwarder) Flush() {
}
func (w *websocketForwarder) Close() error {
	// w.websocket.NextWriter()
	w.mux.Lock()
	defer w.mux.Unlock()
	return w.websocket.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}
func (w *websocketForwarder) CloseConn() error {
	return w.websocket.Close()
}
func (w *websocketForwarder) NextReader() (io.Reader, error) {

	_, reader, err := w.websocket.NextReader()
	if err != nil {
		return nil, err
	}
	return reader, nil
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
