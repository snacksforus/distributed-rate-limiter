package response

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

func Success() *Response {
	return &Response{
		Success: true,
	}
}

func Error(code string, msg string) *Response {
	return &Response{
		Success: false,
		Error: &ResponseError{
			Code:    code,
			Message: msg,
		},
	}
}
