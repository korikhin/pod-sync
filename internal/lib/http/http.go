package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

const (
	HeaderContentType = "Content-Type"
	HeaderRequestID   = "X-Watcher-Request-ID"

	ContentApplicationJSON = "application/json"
)

func ResponseJSON(w http.ResponseWriter, v interface{}, statusCode int) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	if err := enc.Encode(v); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set(HeaderContentType, ContentApplicationJSON)
	w.WriteHeader(statusCode)
	w.Write(buf.Bytes())
}

func DecodeJSON(r io.Reader, v interface{}) error {
	defer io.Copy(io.Discard, r)
	return json.NewDecoder(r).Decode(v)
}
