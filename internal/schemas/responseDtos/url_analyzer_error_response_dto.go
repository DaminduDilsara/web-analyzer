package responseDtos

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
