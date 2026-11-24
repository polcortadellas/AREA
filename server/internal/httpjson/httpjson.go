// Package httpjson writes JSON responses, including the API's error body
// {"error": "<code>"}.
package httpjson

import (
	"encoding/json"
	"net/http"
)

func Write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func Error(w http.ResponseWriter, status int, code string) {
	Write(w, status, map[string]string{"error": code})
}
