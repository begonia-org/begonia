package transport

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"sync/atomic"

	common "github.com/begonia-org/go-sdk/common/api/v1"
)

type ServerSendEventForwarder interface {
	http.ResponseWriter
}

type serverSendEventForwarder struct {
	http.ResponseWriter
	id    int64
	retry int
	event string
}

func NewServerSendEventForwarder(w http.ResponseWriter, req *http.Request, retry int, event string) (ServerSendEventForwarder, error) {
	return &serverSendEventForwarder{w, 0, 3, event}, nil
}
func (s *serverSendEventForwarder) toEventStream(event *common.EventStream) []byte {
	id := fmt.Sprintf("id: %d\n", atomic.LoadInt64(&s.id))
	eventStr := fmt.Sprintf("event: %s\n", event.Event)
	data := fmt.Sprintf("data: %s\n", event.Data)
	retry := fmt.Sprintf("retry: %d\n\n", event.Retry)
	return []byte(id + eventStr + data + retry)
}
func (s *serverSendEventForwarder) Write(message []byte) (int, error) {
	if len(message) == 0 {
		return 0, nil
	}
	event := &common.EventStream{
		Event: s.event,
		Data:  base64.StdEncoding.EncodeToString(message),
		Id:    atomic.LoadInt64(&s.id),
		Retry: int32(s.retry),
	}
	defer atomic.AddInt64(&s.id, 1)
	return s.ResponseWriter.Write(s.toEventStream(event))

}
func (s *serverSendEventForwarder) Flush() {
	s.ResponseWriter.(http.Flusher).Flush()
}
