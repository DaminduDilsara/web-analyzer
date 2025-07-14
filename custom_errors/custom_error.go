package custom_errors

import "fmt"

type CustomError struct {
	Code    int
	Message string
	Err     error
}

func (c CustomError) Error() string {
	return fmt.Sprintf("%s: %v", c.Message, c.Err)
}

func NewCustomError(code int, msg string, err error) *CustomError {
	return &CustomError{
		Code:    code,
		Message: msg,
		Err:     err,
	}
}
