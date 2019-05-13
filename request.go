package accesslog

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

var reqWriteExcludeHeaderDump = map[string]bool{
	"Host":              true, // not in Header map anyway
	"Content-Length":    true,
	"Transfer-Encoding": true,
	"Trailer":           true,
	"Accept-Encoding":   false,
	"Accept-Language":   false,
	"Cache-Control":     false,
	"Connection":        false,
	"Origin":            false,
	"User-Agent":        false,
}

var RequestBodyProcessor = map[string]func(call *Call, req *http.Request){
	MIMEPOSTForm:          readPostForm,
	MIMEJSONPOSTForm:      readPostForm,
	MIMEMultipartPOSTForm: readMultipart,
	MIMEJSON:              requestReadBody,
	MIMEPROTOBUF:          requestReadBody,
	MIMEMSGPACK:           requestReadBody,
	MIMEMSGPACK2:          requestReadBody,
	MIMEHTML:              requestReadBody,
	MIMEXML:               requestReadBody,
	MIMEXML2:              requestReadBody,
	MIMEPlain:             requestReadBody,
}

func requestReadBody(call *Call, req *http.Request) {
	call.RequestBody = *readRequestBody(req)
}

func readRequestHeaders(req *http.Request) map[string]string {
	b := bytes.NewBuffer([]byte(""))
	err := req.Header.WriteSubset(b, reqWriteExcludeHeaderDump)
	if err != nil {
		return map[string]string{}
	}
	headers := map[string]string{}
	for _, header := range strings.Split(b.String(), "\n") {
		values := strings.Split(header, ":")
		if strings.EqualFold(values[0], "") {
			continue
		}
		headers[values[0]] = strings.TrimSpace(values[1])
	}
	return headers
}

func readQueryParams(req *http.Request) map[string]string {
	params := map[string]string{}
	u, err := url.Parse(req.URL.String())
	if err != nil {
		return params
	}
	for k, v := range u.Query() {
		if len(v) < 1 {
			continue
		}
		// TODO: v is a list, and we should be showing a list of values
		// rather than assuming a single value always, gotta change this
		params[k] = v[0]
	}
	return params
}

func readPostForm(call *Call, req *http.Request) {
	call.PostForm = map[string]string{}
	for _, param := range strings.Split(*readRequestBody(req), "&") {
		value := strings.Split(param, "=")
		if len(value) == 2 {
			call.PostForm[value[0]] = value[1]
		}
	}
}

func readMultipart(call *Call, req *http.Request) {
	call.RequestHeader["Content-Type"] = "multipart/form-data"
	_ = req.ParseMultipartForm(MaxInMemoryMultipartSize)
	call.PostForm = map[string]string{}
	for key, val := range req.MultipartForm.Value {
		call.PostForm[key] = val[0]
	}
}

func readRequestBody(req *http.Request) *string {
	var err error
	save := req.Body
	if req.Body == nil {
		req.Body = nil
	} else {
		save, req.Body, err = drainBody(req.Body)
		if err != nil {
			return nil
		}
	}
	if req.Body == nil {
		return nil
	}
	b := bytes.NewBuffer([]byte(""))
	chunked := len(req.TransferEncoding) > 0 && req.TransferEncoding[0] == "chunked"
	var dest io.Writer = b
	if chunked {
		dest = httputil.NewChunkedWriter(dest)
	}
	_, err = io.Copy(dest, req.Body)
	if chunked {
		_ = dest.(io.Closer).Close()
		_, _ = io.WriteString(b, "\r\n")
	}
	req.Body = save
	body := b.String()
	return &body
}
