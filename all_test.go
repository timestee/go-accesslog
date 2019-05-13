package accesslog

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func newMultipartRequest(url string, params map[string]string) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	_ = writer.Close()
	req, err := http.NewRequest("POST", url, body)
	if err == nil {
		req.Header.Add("Content-Type", writer.FormDataContentType())
	}
	return req, err
}

func TestMultipartRequest(t *testing.T) {
	Convey("multipart request", t, func() {
		forms := map[string]string{"key1": "val1", "key2": "val2"}
		path := "/test"
		urlParam := map[string]string{"key3": "val3"}
		params := url.Values{}
		for k, v := range urlParam {
			params.Add(k, v)
		}
		pathWithParam := path + "?" + params.Encode()

		request, err := newMultipartRequest(pathWithParam, forms)
		testHeader := "x-test-header"
		testHeaderVal := "some-random-string"
		request.Header.Add(testHeader, testHeaderVal)

		testResponseBody := map[string]string{"test": "yo"}
		body := new(bytes.Buffer)
		_ = json.NewEncoder(body).Encode(testResponseBody)

		Convey("multipart request new", func() {
			So(err, ShouldBeNil)
			Convey("send multipart request", func() {
				next := func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Header().Set("Content-Type", MIMEJSON)
					_, err := w.Write(body.Bytes())
					So(err, ShouldBeNil)
				}
				call := &Call{}
				outputRecorder := NewResponseRecorder(httptest.NewRecorder(), true)
				Before(call, request)
				next(outputRecorder, request)
				After(call, outputRecorder, request)

				Convey("Call CurrentPath", func() {
					So(call.CurrentPath, ShouldEqual, path)
				})
				Convey("Call MethodType", func() {
					So(call.MethodType, ShouldEqual, request.Method)
				})
				Convey("Call RequestHeader", func() {
					So(call.RequestHeader, ShouldContainKey, "Content-Type")
					mimeKey := textproto.CanonicalMIMEHeaderKey(testHeader)
					So(call.RequestHeader, ShouldContainKey, mimeKey)
					So(call.RequestHeader[mimeKey], ShouldEqual, testHeaderVal)
					So(request.Header.Get("Content-Type"), ShouldContainSubstring, call.RequestHeader["Content-Type"])
				})
				Convey("Call RequestBody", func() {
					So(call.RequestBody, ShouldBeEmpty)
				})
				Convey("Call PostForm", func() {
					So(call.PostForm, ShouldResemble, forms)
				})
				Convey("Call RequestUrlParams", func() {
					So(call.RequestUrlParams, ShouldResemble, urlParam)
				})
				Convey("Call ResponseHeader", func() {
					So(call.ResponseHeader, ShouldContainKey, "Content-Type")
					So(call.ResponseHeader["Content-Type"], ShouldEqual, MIMEJSON)
				})
				Convey("Call ResponseBody", func() {
					So(call.ResponseBody, ShouldEqual, string(body.Bytes()))
				})
				Convey("Call ResponseCode", func() {
					So(call.ResponseCode, ShouldEqual, http.StatusOK)
				})
				Convey("Call Cost", func() {
					So(call.Cost, ShouldNotBeEmpty)
				})
			})
		})
	})

}
