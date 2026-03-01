package response

// Error represents an error from the API
type Error struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// Response represents the response from the API
type Response struct {
	Success bool   `json:"success,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

func Success() *Response {
	return &Response{
		Success: true,
	}
}

func NewError(code string, msg string) *Response {
	return &Response{
		Success: false,
		Error: &Error{
			Code:    code,
			Message: msg,
		},
	}
}
