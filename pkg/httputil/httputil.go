package httputil

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// UnmarshalJSON unmarshals JSON data from the HTTP request body.
func UnmarshalJSON(r *http.Request, v interface{}) error {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}
	return json.Unmarshal(b, v)
}

type httpError struct {
	Error string `json:"error"`
}

// Error creates an HTTP response to handle errors.
func Error(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	Response(w, &httpError{Error: err.Error()})
	log.Printf("[ERROR %d] %q\n", status, err)
}

// Response creates an HTTP response for successful calls.
func Response(w http.ResponseWriter, v interface{}) {
	b, err := json.Marshal(v)
	if err == nil {
		w.Write(b)
	}
}
