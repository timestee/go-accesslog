package accesslog

import (
	"bufio"
	"bytes"
	"errors"
	"net"
	"net/http"
)

type ResponseProxy interface {
	http.ResponseWriter
	Status() int
	ResponseBytes() []byte
}

type responseRecorder struct {
	writer         http.ResponseWriter
	statusCode     int
	recordResponse bool
	Body           *bytes.Buffer
}

func NewResponseRecorder(w http.ResponseWriter, recordResponse bool) ResponseProxy {
	return &responseRecorder{
		writer:         w,
		statusCode:     http.StatusOK,
		recordResponse: recordResponse,
	}
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.writer.WriteHeader(statusCode)
}

func (r *responseRecorder) Header() http.Header {
	return r.writer.Header()
}

func (r *responseRecorder) Write(buf []byte) (int, error) {
	n, err := r.writer.Write(buf)
	if err == nil && r.recordResponse {
		if r.Body == nil {
			r.Body = bytes.NewBuffer(nil)
		}
		r.Body.Write(buf)
	}
	return n, err
}

func (r *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := r.writer.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, errors.New("error in hijacker")
}

func (r *responseRecorder) Status() int {
	return r.statusCode
}

var emptyBytesSlice []byte

func (r *responseRecorder) ResponseBytes() []byte {
	if r.Body == nil {
		return emptyBytesSlice
	}
	return r.Body.Bytes()
}
