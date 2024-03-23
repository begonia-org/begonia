// BEGONIA API Gateway Signature
// based on https://github.com/datastream/aws/blob/master/signv4.go
package signature

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type RequestHeader struct {
	headers map[string]string
}

type GatewayRequest struct {
	Headers *RequestHeader
	Payload io.ReadCloser
	Method  string
	URL     *url.URL
	Host    string
}
type AppAuthSigner interface {
	Sign(request *GatewayRequest) (string, error)
	SignRequest(request *GatewayRequest) error
}

type AppAuthSignerImpl struct {
	Key    string
	Secret string
}

const (
	DateFormat           = "20060102T150405Z"
	SignAlgorithm        = "X-Sign-HMAC-SHA256"
	HeaderXDateTime      = "X-Date"
	HeaderXHost          = "host"
	HeaderXAuthorization = "Authorization"
	HeaderXContentSha256 = "X-Content-Sha256"
	HeaderXAccessKey     = "X-Access-Key"
)

func (h *RequestHeader) Set(key, value string) {
	h.headers[strings.ToLower(key)] = value

}
func (h *RequestHeader) Get(key string) string {
	return h.headers[strings.ToLower(key)]

}
func (h *RequestHeader) Del(key string) {
	delete(h.headers, strings.ToLower(key))
}
func (h *RequestHeader) Keys() []string {
	var keys []string
	for k := range h.headers {
		keys = append(keys, k)
	}
	return keys
}
func (h *RequestHeader) ToMetadata() metadata.MD {
	md := metadata.New(h.headers)
	return md
}
func NewRequestHeader() *RequestHeader {
	return &RequestHeader{headers: make(map[string]string)}
}
func NewGatewayRequestFromGrpc(ctx context.Context, req interface{}, fullMethod string) (*GatewayRequest, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	headers := &RequestHeader{headers: make(map[string]string)}
	host := ""
	uri := fullMethod
	method := fullMethod
	if ok {
		for k, v := range md {
			if strings.EqualFold(k, "x-forwarded-host") {
				host = v[0]
			}
			if strings.EqualFold(k, "uri") {
				uri = v[0]
			}
			if strings.EqualFold(k, "x-http-method") {
				method = v[0]
			}
			values := []string{}
			for _, val := range v {
				hs := strings.Split(val, ",")
				for index, val := range hs {
					hs[index] = strings.TrimSpace(val)
				}
				values = append(values, hs...)
			}
			headers.Set(k, strings.Join(values, ","))
		}
	}
	u, _ := url.Parse(fmt.Sprintf("http://%s%s", host, uri))
	// 兼容 application/json
	payload, _ := protojson.Marshal(req.(proto.Message))
	// // log.Printf("request payloadxxxx:%v", payload)
	return &GatewayRequest{Headers: headers,
		Method:  method,
		Host:    host,
		URL:     u,
		Payload: io.NopCloser(bytes.NewBuffer(payload)),
	}, nil
}
func NewGatewayRequestFromHttp(req *http.Request) (*GatewayRequest, error) {
	// headers := make(map[string]string)
	headers := &RequestHeader{headers: make(map[string]string)}
	for k, v := range req.Header {
		headers.Set(k, strings.Join(v, ","))
	}
	payload, _ := io.ReadAll(req.Body)
	req.Body = io.NopCloser(bytes.NewBuffer(payload))
	return &GatewayRequest{Headers: headers, Method: req.Method, URL: req.URL, Host: req.Host, Payload: io.NopCloser(bytes.NewBuffer(payload))}, nil
}
func NewAppAuthSigner(key, secret string) AppAuthSigner {
	return &AppAuthSignerImpl{Key: key, Secret: secret}
}
func (app *AppAuthSignerImpl) hmacsha256(keyByte []byte, dataStr string) ([]byte, error) {
	hm := hmac.New(sha256.New, []byte(keyByte))
	if _, err := hm.Write([]byte(dataStr)); err != nil {
		return nil, err
	}
	return hm.Sum(nil), nil
}

// Build a CanonicalRequest from a regular request string
func (app *AppAuthSignerImpl) CanonicalRequest(request *GatewayRequest, signedHeaders []string) (string, error) {
	var hexencode string
	var err error
	if hex := request.Headers.Get(HeaderXContentSha256); hex != "" {
		hexencode = hex
	} else {
		bodyData, err := app.RequestPayload(request)
		if err != nil {
			return "", err
		}
		hexencode, err = app.HexEncodeSHA256Hash(bodyData)
		if err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s", strings.ToUpper(request.Method), app.CanonicalURI(request), app.CanonicalQueryString(request), app.CanonicalHeaders(request, signedHeaders), strings.Join(signedHeaders, ";"), hexencode), err
}

// CanonicalURI returns request uri
func (app *AppAuthSignerImpl) CanonicalURI(request *GatewayRequest) string {
	pattens := strings.Split(request.URL.Path, "/")
	var uriSlice []string
	for _, v := range pattens {
		uriSlice = append(uriSlice, url.PathEscape(v))
	}
	urlpath := strings.Join(uriSlice, "/")
	if len(urlpath) == 0 || urlpath[len(urlpath)-1] != '/' {
		urlpath = urlpath + "/"
	}
	// // log.Printf("canonicalURI:%s", urlpath)
	return urlpath
}

// CanonicalQueryString
func (app *AppAuthSignerImpl) CanonicalQueryString(request *GatewayRequest) string {
	var keys []string
	queryMap := request.URL.Query()
	for key := range queryMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var query []string
	for _, key := range keys {
		k := escape(key)
		sort.Strings(queryMap[key])
		for _, v := range queryMap[key] {
			kv := fmt.Sprintf("%s=%s", k, escape(v))
			query = append(query, kv)
		}
	}
	queryStr := strings.Join(query, "&")
	request.URL.RawQuery = queryStr
	// // log.Printf("canonicalQueryString:%s", queryStr)
	return queryStr
}

// CanonicalHeaders
func (app *AppAuthSignerImpl) CanonicalHeaders(request *GatewayRequest, signerHeaders []string) string {
	var canonicalHeaders []string
	header := make(map[string][]string)
	for k, v := range request.Headers.headers {
		val := strings.Split(v, ",")
		header[strings.ToLower(k)] = val
	}
	for _, key := range signerHeaders {
		value := header[strings.ToLower(key)]
		if strings.EqualFold(key, HeaderXHost) {
			value = []string{request.Host}
		}

		sort.Strings(value)
		// log.Println("sort value:", value)
		canonicalHeaders = append(canonicalHeaders, key+":"+strings.TrimSpace(strings.Join(value, ",")))

	}
	return strings.Join(canonicalHeaders, "\n")
}

// SignedHeaders
func (app *AppAuthSignerImpl) SignedHeaders(r *GatewayRequest) []string {
	var signedHeaders []string
	for key := range r.Headers.headers {
		if strings.EqualFold(key, HeaderXAuthorization) {
			return app.getSignaturehHeader(r.Headers.Get(HeaderXAuthorization))
		}
		if r.Headers.headers[key] == "" {
			continue
		}
		signedHeaders = append(signedHeaders, strings.ToLower(key))
	}
	sort.Strings(signedHeaders)
	return signedHeaders
}

// RequestPayload
func (app *AppAuthSignerImpl) RequestPayload(request *GatewayRequest) ([]byte, error) {
	if request.Payload == nil {
		return []byte(""), nil
	}
	bodyByte, err := io.ReadAll(request.Payload)
	if err != nil {
		return []byte(""), err
	}
	request.Payload = io.NopCloser(bytes.NewBuffer(bodyByte))
	return bodyByte, err
}

// Create a "String to Sign".
func (app *AppAuthSignerImpl) StringToSign(canonicalRequest string, t time.Time) (string, error) {
	hashStruct := sha256.New()
	// log.Println("canonicalRequest to sign:", canonicalRequest)
	_, err := hashStruct.Write([]byte(canonicalRequest))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s\n%s\n%x",
		SignAlgorithm, t.UTC().Format(DateFormat), hashStruct.Sum(nil)), nil
}

// Create the  Signature.
func (app *AppAuthSignerImpl) SignStringToSign(stringToSign string, signingKey []byte) (string, error) {
	hmsha, err := app.hmacsha256(signingKey, stringToSign)
	return fmt.Sprintf("%x", hmsha), err
}

// HexEncodeSHA256Hash returns hexcode of sha256
func (app *AppAuthSignerImpl) HexEncodeSHA256Hash(body []byte) (string, error) {
	hashStruct := sha256.New()
	if len(body) == 0 {
		body = []byte("")
	}
	_, err := hashStruct.Write(body)
	return fmt.Sprintf("%x", hashStruct.Sum(nil)), err
}

// Get the finalized value for the "Authorization" header. The signature parameter is the output from SignStringToSign
func (app *AppAuthSignerImpl) AuthHeaderValue(signatureStr, accessKeyStr string, signedHeaders []string) string {
	return fmt.Sprintf("%s Access=%s, SignedHeaders=%s, Signature=%s", SignAlgorithm, accessKeyStr, strings.Join(signedHeaders, ";"), signatureStr)
}

// SignRequest set Authorization header
func (app *AppAuthSignerImpl) Sign(request *GatewayRequest) (string, error) {
	var t time.Time
	var err error
	var date string
	if date := request.Headers.Get(HeaderXDateTime); date != "" {
		t, err = time.Parse(DateFormat, date)
		if time.Since(t) > time.Minute*1 {
			return "", fmt.Errorf("X-Date is expired")
		}
	}
	if err != nil || date == "" {
		t = time.Now()
		request.Headers.Set(HeaderXDateTime, t.UTC().Format(DateFormat))
	}
	request.Headers.Set(HeaderXAccessKey, app.Key)
	signedHeaders := app.SignedHeaders(request)
	// log.Printf("signedHeaders:%v", signedHeaders)

	canonicalRequest, err := app.CanonicalRequest(request, signedHeaders)
	// log.Printf("canonicalRequest:%s*********", canonicalRequest)
	if err != nil {
		return "", fmt.Errorf("Failed to create canonical request: %w", err)
	}
	stringToSignStr, err := app.StringToSign(canonicalRequest, t)
	// log.Printf("stringToSign:%s", stringToSignStr)
	if err != nil {
		return "", fmt.Errorf("Failed to create string to sign: %w", err)
	}
	signatureStr, err := app.SignStringToSign(stringToSignStr, []byte(app.Secret))
	// log.Printf("signature:%s", signatureStr)
	if err != nil {
		return "", fmt.Errorf("Failed to create signature: %w", err)
	}

	return signatureStr, nil
}
func (app *AppAuthSignerImpl) getSignaturehHeader(auth string) []string {
	strArr := strings.Split(auth, ",")
	for _, v := range strArr {
		if strings.Contains(strings.ToLower(v), "signedheaders") {
			// // // log.Printf("origin signedHeaders:%s", v)
			signature := strings.Split(v, "=")
			return strings.Split(signature[1], ";")
		}
	}
	return nil
}
func (app *AppAuthSignerImpl) SignRequest(request *GatewayRequest) error {
	signature, err := app.Sign(request)
	if err != nil {
		return err
	}
	request.Headers.Set(HeaderXAuthorization, app.AuthHeaderValue(signature, app.Key, app.SignedHeaders(request)))
	return nil
}
