// https://datatracker.ietf.org/doc/html/rfc6750
package jwt

import (
	jwtgorequest "github.com/golang-jwt/jwt/v4/request"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"strings"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
}

type HeaderExtractor struct {
	// Bearer
	TokenType string
	jwtgorequest.HeaderExtractor
}

func (e HeaderExtractor) ExtractToken(req *http.Request) (string, error) {
	value, err := e.HeaderExtractor.ExtractToken(req)
	if err != nil {
		return "", nil
	}
	if !strings.HasPrefix(value, e.TokenType) {
		return "", errors.Errorf("token type is not registered type: %s", e.TokenType)
	}
	return value, nil
}
