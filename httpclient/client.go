package httpclient

import (
	"bytes"
	jsonlib "encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/uddmorningsun/go-livingkit"
	"io"
	"net/http"
	"net/http/httputil"
	urllib "net/url"
	"os"
	"strings"
	"time"
)

type HTTPClient struct {
	client       *http.Client
	scheme, host string
	// Scheme://Host:Port or Scheme://Host
	Address string
}

type errServerResponse struct {
	Success  bool           `json:"success"`
	Message  string         `json:"message"`
	Response *http.Response `json:"-"`
}

func (esr *errServerResponse) setFieldValue(success bool, message string, response *http.Response) *errServerResponse {
	esr.Success = success
	esr.Message = message
	esr.Response = response
	return esr
}

// LowerLevelClientOption is a customizable option for initialize lower level http.Client.
type LowerLevelClientOption func(*http.Client) error

// RequestOption is a customizable option for initialize origin http.Request.
type RequestOption func(*http.Request) error

// Option is a customizable option for initialize HTTPClient, details also refer to: WithAddress, etc.
type Option func(*HTTPClient) error

var (
	defaultHeaders = map[string]string{
		"Accept":     "*/*",
		"Connection": "keep-alive",
	}
)

// WithAddress can overwrite the client request address with specified one address.
func WithAddress(address string) Option {
	return func(hc *HTTPClient) error {
		u, err := urllib.Parse(address)
		if err != nil {
			return fmt.Errorf("unable to parse address, error: %s", err)
		}
		// Default `http.Transport` only support http/https scheme, see net/http/transport.go:roundTrip()
		if u.Scheme != "http" && u.Scheme != "https" {
			return fmt.Errorf("invalid scheme only support http(s) scheme")
		}
		if u.Host == "" {
			return fmt.Errorf("no host found in request URL")
		}
		hc.host = u.Host
		hc.scheme = u.Scheme
		hc.Address = address
		return nil
	}
}

// WithLowerLevelClientOption can initialize lower level http.Client.
func WithLowerLevelClientOption(opt LowerLevelClientOption) Option {
	return func(hc *HTTPClient) error {
		if err := opt(hc.client); err != nil {
			return fmt.Errorf("unable to apply lower level client option, error: %s", err)
		}
		return nil
	}
}

// NewHTTPClientWithOptions will initialize HTTPClient with series of Option, design inspired by docker/docker.
func NewHTTPClientWithOptions(opts ...Option) (*HTTPClient, error) {
	hc := &HTTPClient{
		client: http.DefaultClient,
	}
	for _, opt := range opts {
		if err := opt(hc); err != nil {
			return nil, fmt.Errorf("unable apply option, error: %s", err)
		}
	}
	return hc, nil
}

func (hc *HTTPClient) newRequest(method, path string, body io.Reader) (*http.Request, error) {
	logrus.Debugf("request address: %s, path: %s, method: %s", hc.Address, path, method)
	expectedPayload := method == http.MethodPost || method == http.MethodPatch || method == http.MethodPut
	if expectedPayload && body == nil {
		body = bytes.NewReader(nil)
	}
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize go origin http request, error: %s", err)
	}
	for key, value := range defaultHeaders {
		req.Header.Set(key, value)
	}
	if expectedPayload && req.Header.Get(livingkit.ContentType) == "" {
		req.Header.Set(livingkit.ContentType, livingkit.TextPlain)
	}
	return req, nil
}

// DoRequest will do real request.
func (hc *HTTPClient) DoRequest(method, path string, body io.Reader, reqOpts ...RequestOption) (*http.Response, error) {
	req, err := hc.newRequest(method, path, body)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize request, error: %s", err)
	}
	for _, opt := range reqOpts {
		if err := opt(req); err != nil {
			return nil, fmt.Errorf("unable to apply request option, error: %s", err)
		}
	}
	startedTime := time.Now()
	response, err := hc.client.Do(req)
	logrus.Debugf("request elapsed time: %s", time.Since(startedTime).String())
	if err != nil {
		return nil, fmt.Errorf("request (%s:%s) failed, error: %s", path, method, err)
	}
	DumpVerboseRequestResponse(os.Stderr, req, response)
	return response, nil
}

// OK checks if status code of the response is between [200, 400), this will return true if OK.
func (hc *HTTPClient) OK(resp *http.Response) bool {
	if http.StatusOK <= resp.StatusCode && resp.StatusCode < http.StatusBadRequest {
		return true
	}
	return false
}

// HandleResponse converts http.Response to given entity. If HTTP code is not [200, 400), it will
func (hc *HTTPClient) HandleResponse(resp *http.Response, entity interface{}) *errServerResponse {
	defer func() {
		if !resp.Close {
			resp.Body.Close()
		}
	}()

	var esr = new(errServerResponse)
	b := new(bytes.Buffer)
	if _, err := b.ReadFrom(resp.Body); err != nil {
		return esr.setFieldValue(false, fmt.Sprintf("unable to read response, error: %s", err), resp)
	}
	if hc.OK(resp) {
		if err := jsonlib.Unmarshal(b.Bytes(), &entity); err != nil {
			return esr.setFieldValue(false, fmt.Sprintf("unable to unmarshal json response, error: %s", err), resp)
		}
		return nil
	}
	if err := jsonlib.Unmarshal(b.Bytes(), &esr); err != nil {
		return esr.setFieldValue(false, fmt.Sprintf("unable to unmarshal json error response, error: %s", err), resp)
	}
	esr.Response = resp
	return esr
}

// PrepareBody prepares the given HTTP body data for POST/PUT/PATCH generally, refer to requests/models.py:PrepareRequest.prepare_body.
// https://learning.postman.com/docs/sending-requests/requests/#sending-body-data
func (hc *HTTPClient) PrepareBody(json interface{}, data urllib.Values) (io.Reader, string, error) {
	var (
		body        io.Reader
		contentType string
	)
	if data == nil && json != nil {
		payload, err := jsonlib.Marshal(json)
		if err != nil {
			return nil, "", fmt.Errorf("invalid json data, error: %s", err)
		}
		body = bytes.NewReader(payload)
		contentType = livingkit.ApplicationJSON
	} else if data != nil {
		payload := data.Encode()
		body = bytes.NewReader([]byte(payload))
		contentType = livingkit.ApplicationXWWWFormUrlencoded
	}
	return body, contentType, nil
}

// PreparePath prepares the given path and query string to new api. If origin path contains query params, new params will append.
func (hc *HTTPClient) PreparePath(path string, params urllib.Values) (string, error) {
	if params == nil {
		return path, nil
	}
	u, err := urllib.Parse(path)
	if err != nil {
		return "", fmt.Errorf("unable to prepare path: %s", err)
	}
	if u.RawQuery != "" {
		u.RawQuery = fmt.Sprintf("%s&%s", u.RawQuery, params.Encode())
	} else {
		u.RawQuery = params.Encode()
	}
	return u.String(), nil
}

func (hc *HTTPClient) commonPostPatchPut(method, path string, json interface{}, data urllib.Values, opts ...RequestOption) (*http.Response, error) {
	switch method {
	case http.MethodGet, http.MethodDelete:
		logrus.Warningf("method: %s recommends that should not body", method)
	}
	body, ct, err := hc.PrepareBody(json, data)
	if err != nil {
		return nil, fmt.Errorf("prepare body failed, error: %s", err)
	}
	opts = append(opts, func(request *http.Request) error {
		request.Header.Set(livingkit.ContentType, ct)
		return nil
	})
	return hc.DoRequest(method, path, body, opts...)
}

// Post sends a http.MethodPost request.
func (hc *HTTPClient) Post(path string, json interface{}, data urllib.Values, opts ...RequestOption) (*http.Response, error) {
	return hc.commonPostPatchPut(http.MethodPost, path, json, data, opts...)
}

// Patch sends a http.MethodPatch request.
func (hc *HTTPClient) Patch(path string, json interface{}, data urllib.Values, opts ...RequestOption) (*http.Response, error) {
	return hc.commonPostPatchPut(http.MethodPatch, path, json, data, opts...)
}

func (hc *HTTPClient) commonGetDelete(method, path string, params urllib.Values, opts ...RequestOption) (*http.Response, error) {
	switch {
	case method != http.MethodGet, method != http.MethodDelete:
		logrus.Warningf("method: %s recommends that should not URL query params", method)
	}
	path, err := hc.PreparePath(path, params)
	if err != nil {
		return nil, fmt.Errorf("unable to prepare path: %s", path)
	}
	return hc.DoRequest(method, path, nil, opts...)
}

// Get sends a http.MethodGet request.
func (hc *HTTPClient) Get(path string, params urllib.Values, opts ...RequestOption) (*http.Response, error) {
	return hc.commonGetDelete(http.MethodGet, path, params, opts...)
}

// Delete sends a http.MethodDelete request.
func (hc *HTTPClient) Delete(path string, params urllib.Values, opts ...RequestOption) (*http.Response, error) {
	return hc.commonGetDelete(http.MethodDelete, path, params, opts...)
}

// DumpVerboseRequestResponse returns the given request and response in its HTTP/1.x wire representation.
func DumpVerboseRequestResponse(w io.Writer, req *http.Request, resp *http.Response) {
	includeBody := os.Getenv(livingkit.DebugHTTPClientBody) != ""
	if os.Getenv(livingkit.DebugHTTPClient) == "" {
		return
	}
	dump, err := httputil.DumpRequest(req, includeBody)
	if err == nil {
		fmt.Fprintln(w, strings.Repeat(">", 100))
		bytes.NewBuffer(dump).WriteTo(w)
		if includeBody {
			fmt.Println()
		}
	}
	dump, err = httputil.DumpResponse(resp, includeBody)
	if err == nil {
		fmt.Fprintln(w, strings.Repeat("<", 100))
		bytes.NewBuffer(dump).WriteTo(w)
	}
}
