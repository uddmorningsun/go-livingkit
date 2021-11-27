package ginlib

import (
	"net/http"
)

var (
	ErrInvalidRequestParams = NewErrorCode(1000, http.StatusBadRequest, "invalid request params")
	ErrUnknownError = NewErrorCode(9999, http.StatusInternalServerError, "unknown server internal error")
)
