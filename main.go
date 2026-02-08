package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// ResponseError represents an error from the API
type ResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Response represents the response from the API
type Response struct {
	Success bool           `json:"success"`
	Error   *ResponseError `json:"error"`
}

func main() {
	// API has a single endpoint that just returns success.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		resp := Response{
			Success: true,
		}
		data, _ := json.Marshal(resp)
		_, _ = w.Write(data)
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
