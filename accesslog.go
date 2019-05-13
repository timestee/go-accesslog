package accesslog

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// TODO omitempty with empty map
type Call struct {
	CurrentPath      string            `json:"call_path"`
	MethodType       string            `json:"method_type"`
	RequestBody      string            `json:"request_body,omitempty"`
	RequestHeader    map[string]string `json:"request_header"`
	PostForm         map[string]string `json:"post_form,omitempty"`
	RequestUrlParams map[string]string `json:"request_url_params,omitempty"`

	ResponseCode   int               `json:"response_code"`
	ResponseBody   string            `json:"response_body"`
	ResponseHeader map[string]string `json:"response_header"`

	Cost  string `json:"cost"`
	start time.Time
}

func Before(call *Call, req *http.Request) {
	call.start = time.Now()
	call.MethodType = req.Method
	call.CurrentPath = strings.Split(req.URL.String(), "?")[0]
	call.RequestHeader = readRequestHeaders(req)
	call.RequestUrlParams = readQueryParams(req)

	ct, ok := call.RequestHeader["Content-Type"]
	if !ok {
		return
	}
	ct = strings.TrimSpace(ct)
	handler, ok := RequestBodyProcessor[ct]
	if !ok && strings.Contains(ct, MIMEMultipartPOSTForm) {
		handler = readMultipart
	}
	if handler == nil {
		handler = requestReadBody
	}
	handler(call, req)
}

func After(call *Call, record ResponseProxy, r *http.Request) {
	if strings.Contains(r.RequestURI, ".ico") {
		return
	}
	call.ResponseCode = record.Status()
	call.ResponseHeader = readResponseHeaders(record.Header())
	call.ResponseBody = string(record.ResponseBytes())
	call.Cost = fmt.Sprintf("%s", time.Since(call.start))
}
