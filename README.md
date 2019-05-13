# go-accesslog [![Build Status](https://travis-ci.org/timestee/go-accesslog.svg?branch=master)](https://travis-ci.org/timestee/go-accesslog) [![Go Walker](https://gowalker.org/api/v1/badge)](https://gowalker.org/github.com/timestee/go-accesslog)  [![GoDoc](https://godoc.org/github.com/timestee/go-accesslog?status.svg)](https://godoc.org/github.com/timestee/go-accesslog)

A simple HTTP access log for golang.

Usage:

```golang
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
```
