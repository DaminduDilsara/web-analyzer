package response_dtos

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
