package accesslog

import (
	"net/http"
	"strings"
)

func readResponseHeaders(header http.Header) map[string]string {
	headers := map[string]string{}
	for k, v := range header {
		headers[k] = strings.Join(v, " ")
	}
	return headers
}
