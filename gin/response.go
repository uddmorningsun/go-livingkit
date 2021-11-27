package ginlib

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

// ResponseOK will write valid JSON response.
func ResponseOK(c *gin.Context, httpCode int, data interface{}) {
	if data == nil {
		data = map[string]string{}
	}
	c.JSON(httpCode, data)
}

// ResponseError will stop call the remaining handlers and return specific JSON error response.
// You can customize response format if param `jsonObj` concrete value type is not constructed by NewErrorCode function.
// NOTE: if you call ResponseError in handler function, don't forget to add `return` clause to interrupt function call.
func ResponseError(c *gin.Context, jsonObj interface{}, httpCode ...int) {
	code, ok := jsonObj.(ErrorCode)
	if ok {
		if code.err != nil {
			fe, ok := code.err.(validator.ValidationErrors)
			if gin.IsDebugging() && ok {
				for _, value := range fe {
					logrus.Errorf(
						"parameter validation error: SNS: %s || ParamValue: %q || DefinedTag: %s(%s) || Type: %s",
						value.ActualTag(), value.Param(), value.Tag(), value.ActualTag(), value.Tag(),
					)
				}
			}
			logrus.Errorf("Error on [%s - %s]: %+v", c.FullPath(), c.Request.Method, code.err)
			// If caller wrappers error(e.g.: `fmt.Errorf("%w", err)`), we will try to extract underlying error.
			if err := errors.Unwrap(code.err); err != nil {
				logrus.Errorf("Underlying error: %+v", err)
			}
		}
		c.AbortWithStatusJSON(code.httpCode, code)
		return
	}
	if len(httpCode) == 0 {
		panic("go-livingkit/usage: required one http code param at least if you want to customize error response format")
	}
	c.AbortWithStatusJSON(httpCode[0], jsonObj)
}

// ErrorCode is uniform error response struct definition, user should not initialize new errorcode with it.
// If you want to define errorcode, recommends call NewErrorCode directly.
type ErrorCode struct {
	code, httpCode     int
	message, delimiter string
	err                error
}

// NewErrorCode will new specific error code with customizable code number, HTTPCode and message.
// Default delimiter is blank space.
func NewErrorCode(code, httpCode int, message string) ErrorCode {
	return ErrorCode{code: code, httpCode: httpCode, message: message, delimiter: " "}
}

// SetDelimiter supports overwrite default blank space delimiter.
func (ec ErrorCode) SetDelimiter(delimiter string) ErrorCode {
	ec.delimiter = delimiter
	return ec
}

// WithMessage supports append or overwrite a customized message.
func (ec ErrorCode) WithMessage(message string, replace bool) ErrorCode {
	if replace {
		ec.message = message
	} else {
		ec.message = fmt.Sprintf("%s%s%s", ec.message, ec.delimiter, message)
	}
	return ec
}

// WithError supports wrap error, error could be origin error or call `fmt.Errorf("%w", err)`.
func (ec ErrorCode) WithError(myerr error) ErrorCode {
	if myerr != nil {
		ec.err = myerr
	}
	return ec
}

func (ec ErrorCode) String() string {
	return ec.message
}
