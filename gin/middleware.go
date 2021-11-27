package ginlib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/uddmorningsun/go-livingkit"
	"io"
	"net/http"
	"strings"
)

// IsValidUUIDParam checks the value of the URL params whether is valid UUID string or not.
func IsValidUUIDParam(param string) gin.HandlerFunc {
	return func(c *gin.Context) {
		value := c.Param(param)
		if value == "" {
			ResponseError(c, ErrInvalidRequestParams)
			return
		}
		if !livingkit.IsValidUUID(value) {
			logrus.Errorf("URL param: %s is invalid UUID format", param)
			ResponseError(c, ErrInvalidRequestParams)
			return
		}
		c.Next()
	}
}

// DebugLogRequestData prints request payload and URL params for debugging.
func DebugLogRequestData(out io.Writer) gin.HandlerFunc {
	if out == nil {
		out = gin.DefaultErrorWriter
	}
	return func(c *gin.Context) {
		if !gin.IsDebugging() || logrus.GetLevel() != logrus.DebugLevel {
			return
		}
		var (
			data                interface{}
			method              = c.Request.Method
			contentType         = c.GetHeader(livingkit.ContentType)
			equalSignDelimiters = strings.Repeat("=", 50)
		)
		switch {
		case method == http.MethodGet || method == http.MethodDelete:
			data = c.Request.URL.Query()
		case method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch:
			// Here we must copy request body to debug print payload. If not, unmarshal will drain body and body will be empty,
			// and then will produce a EOF error.
			// However, Gin `Context.ShouldBindBodyWith()` make a body cache in context with `c.Set()` and `c.Get()` based gin.BodyBytesKey.
			// If call this method for debugging in here, framework caller should also call it.
			// `net/http: Request.Clone()` is not deep copy, see: https://github.com/golang/go/issues/36095
			var buf bytes.Buffer
			if _, err := buf.ReadFrom(c.Request.Body); err != nil {
				logrus.Warningf("read body data error: %s", err)
				c.Next()
				return
			}
			if strings.HasPrefix(contentType, livingkit.ApplicationJSON) {
				if err := json.Unmarshal(buf.Bytes(), &data); err != nil {
					logrus.Warningf("invalid JSON format error: %s", err)
					c.Next()
					return
				}
			} else if strings.HasPrefix(contentType, livingkit.MultipartFormData) {
				c.Request.Body = io.NopCloser(bytes.NewReader(buf.Bytes()))
				form, err := c.MultipartForm()
				if err != nil {
					logrus.Warningf("parse multipart form error: %s", err)
					c.Next()
					return
				}
				data = form.Value
				if form.File != nil {
					_, _ = fmt.Fprintf(out, "%s: find file parts: %v\n", equalSignDelimiters, form.File)
				}
			}
			c.Request.Body = io.NopCloser(bytes.NewReader(buf.Bytes()))
		}
		_, _ = fmt.Fprintf(
			out,
			"%s\n[%s - %s - %s] [%s - %s] \n%s\n",
			equalSignDelimiters,
			c.FullPath(), method, c.ClientIP(), c.Request.Header, data,
			equalSignDelimiters,
		)
		c.Next()
	}
}

// RecoverJSONResponse is similar with Python Flask `Flask.handler_exception()`, it will always return JSON response if panic.
func RecoverJSONResponse(out io.Writer) gin.HandlerFunc {
	if out == nil {
		out = gin.DefaultErrorWriter
	}
	return gin.CustomRecoveryWithWriter(out, func(c *gin.Context, err interface{}) {
		switch value := err.(type) {
		// Parameter validation with Gin `binding` tag based on `go-playground/validator`, application caller can be used:
		// 	if err := c.ShouldBindQuery(&param); err != nil {
		// 		panic(err)
		// 		// Or
		// 		panic(ErrInvalidRequestParams.WithError(err))
		// 	}
		case validator.ValidationErrors:
			ResponseError(c, ErrInvalidRequestParams.WithError(value))
		case ErrorCode:
			ResponseError(c, value)
		default:
			underlyingErr, ok := value.(interface {
				Error() string
			})
			if ok {
				ResponseError(c, ErrUnknownError.WithError(underlyingErr))
			} else {
				ResponseError(c, ErrUnknownError.WithError(fmt.Errorf("%v", value)))
			}
		}
	})
}
