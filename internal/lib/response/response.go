package response

type Response struct {
	Status  string `json:"status" env-required:"true"`
	Message string `json:"message" env-required:"true"`
}

func OK(msg string) *Response {
	return &Response{
		Status:  "OK",
		Message: msg,
	}
}

func Error(err error) *Response {
	return &Response{
		Status:  "Error",
		Message: err.Error(),
	}
}
