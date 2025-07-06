package app_errors

import "errors"

var ErrorBadRequest error
var ErrorInternal error

func init() {
	ErrorBadRequest = errors.New("bad request")
	ErrorInternal = errors.New("error internal")
}
