package transport

import "net/http"

type FastHttpHeaderAdapter struct {
}

type FastHttpResponseAdapter struct {
}

type FastHttpRequestAdapter struct{}

func (r *FastHttpResponseAdapter) Header() http.Header {
	return nil
}
