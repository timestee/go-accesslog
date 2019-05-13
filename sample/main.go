package main

import (
	"bytes"
	"encoding/json"
	"github.com/timestee/go-accesslog"
	"log"
	"net/http"
)

type Middleware func(next http.Handler) http.Handler

func AccessLog(cb func(call *accesslog.Call)) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			// before request,install ResponseWriter proxy
			call := &accesslog.Call{}
			rwProxy := accesslog.NewResponseRecorder(rw, true)
			accesslog.Before(call, r)

			next.ServeHTTP(rwProxy, r)

			// after request and callback
			accesslog.After(call, rwProxy, r)
			cb(call)
		})
	}
}

// Use applies a list of middlewares onto a HandlerFunc
func Use(handler http.HandlerFunc, m ...Middleware) http.Handler {
	mc := len(m) - 1
	for i := range m {
		m := m[mc-i]
		handler = m(handler).ServeHTTP
	}
	return handler
}

type builder struct {
	*http.ServeMux
	middlewares []Middleware
}

// http.Handler interface
func (p *builder) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	handler, _ := p.ServeMux.Handler(req)
	Use(handler.ServeHTTP, p.middlewares...).ServeHTTP(res, req)
}

func New(middlewares ...Middleware) *builder {
	return &builder{ServeMux: http.NewServeMux(), middlewares: middlewares}
}

func main() {
	b := New(AccessLog(func(call *accesslog.Call) {
		body := new(bytes.Buffer)
		_ = json.NewEncoder(body).Encode(call)
		log.Println(body)
	}))
	b.ServeMux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(200)
		testResponseBody := map[string]string{"test": "yo"}
		body := new(bytes.Buffer)
		_ = json.NewEncoder(body).Encode(testResponseBody)
		w.Write(body.Bytes())
	})
	http.ListenAndServe(":12001", b)
}
