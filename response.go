package main

import (
	"encoding/json"
	"net/http"
)

type response struct {
	Msg   string `json:"msg,omitempty"`
	Error error  `json:"error,omitempty"`
}

// Send API response back to client
func (r response) Send(httpStatus int, w http.ResponseWriter) {
	json, _ := json.Marshal(r)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(httpStatus)
	w.Write(json)
}
