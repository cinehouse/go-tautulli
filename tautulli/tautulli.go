package tautulli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/go-querystring/query"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
)

const (
	defaultAPIPath = "/api/v2"
	userAgent      = "go-tautulli"
)

var errNonNilContext = errors.New("context must be non-nil")

// A Client manages communication with the Tautulli API.
type Client struct {
	// HTTP client used to communicate with the API.
	client *http.Client
	// Base URL for API requests. BaseURL should always be specified with a trailing slash.
	BaseURL *url.URL
	// API Key for the Tautulli API
	APIKey string
	// User agent used when communicating with the Tautulli API.
	UserAgent string

	// Reuse a single struct instead of allocating one for each service on the heap.
	common service

	// Services used for talking to different parts of the Tautulli API.
	Notifications *NotificationsService

	// Debug mode. Add
	Debug bool
}

type service struct {
	client *Client
}

// Client returns the http.Client used by this Tautulli client.
func (c *Client) Client() *http.Client {
	clientCopy := *c.client
	return &clientCopy
}

// CommonParameters are the parameters that are common to all requests.
type CommonParameters struct {
	APIKey  string `url:"apikey"`             // API key for the Tautulli API
	OutType string `url:"out_type,omitempty"` // Output format of the response
	Debug   int    `url:"debug,omitempty"`    // Debug mode
}

// encodeParameters encodes parameters in a form suitable for a URL query.
// params must be a struct whose fields may contain "url" tags.
func encodeParameters(params interface{}) (string, error) {
	v := reflect.ValueOf(params)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return "", nil
	}

	qs, err := query.Values(params)
	if err != nil {
		return "", err
	}

	return qs.Encode(), nil
}

type ClientOptions struct {
	// APIPath for API requests.
	APIPath string
	// Debug mode. Add additional logging to the client.
	Debug bool
}

// NewClient returns a new Tautulli API client. If a nil httpClient is
// provided, a new http.Client will be used.
func NewClient(httpClient *http.Client, baseURL *url.URL, apiKey string, options *ClientOptions) *Client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	if baseURL == nil {
		baseURL = &url.URL{}
	}
	debug := false
	if options != nil {
		if options.APIPath != "" {
			baseURL, _ = baseURL.Parse(options.APIPath)
		}
		if options.Debug {
			debug = true
		}
	}
	u, _ := url.Parse(baseURL.String() + defaultAPIPath)
	c := &Client{client: httpClient, BaseURL: u, APIKey: apiKey, UserAgent: userAgent, Debug: debug}
	c.common.client = c
	c.Notifications = (*NotificationsService)(&c.common)
	return c
}

// NewRequest creates an API request. If specified, the value pointed to
// by params is URL encoded and included as the query parameters.
func (c *Client) NewRequest(method, urlStr string) (*http.Request, error) {
	if c.Debug {
		log.Printf("New request: %s %s", method, urlStr)
	}
	apiDebug := 0
	if c.Debug {
		apiDebug = 1
	}
	commonParams, err := encodeParameters(&CommonParameters{
		APIKey:  c.APIKey,
		OutType: "json",
		Debug:   apiDebug,
	})
	if err != nil {
		return nil, err
	}
	if c.Debug {
		log.Printf("Common parameters: %v", commonParams)
	}

	u := c.BaseURL
	u.RawQuery = urlStr + "&" + commonParams
	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return nil, err
	}

	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	if c.Debug {
		log.Printf("HTTP Request: %v", req)
	}
	return req, nil
}

// Response is a Tautulli API response. This wraps the standard http.Response
// returned from Tautulli and provides convenient access to things like
// pagination links.
type Response struct {
	*http.Response
}

// newResponse creates a new Response for the provided http.Response.
// r must not be nil.
func newResponse(r *http.Response) *Response {
	response := &Response{Response: r}
	return response
}

// BareDo sends an API request and lets you handle the api response. If an error
// or API Error occurs, the error will contain more information. Otherwise you
// are supposed to read and close the response's Body. If rate limit is exceeded
// and reset time is in the future, BareDo returns *RateLimitError immediately
// without making a network API call.
//
// The provided ctx must be non-nil, if it is nil an error is returned. If it is
// canceled or times out, ctx.Err() will be returned.
func (c *Client) BareDo(ctx context.Context, req *http.Request) (*Response, error) {
	if ctx == nil {
		return nil, errNonNilContext
	}

	req = req.WithContext(ctx)

	resp, err := c.client.Do(req)
	if err != nil {
		// If we got an error, and the context has been canceled,
		// the context's error is probably more useful.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// If the error type is *url.Error, sanitize its URL before returning.
		if e, ok := err.(*url.Error); ok {
			if url, err := url.Parse(e.URL); err == nil {
				e.URL = sanitizeURL(url).String()
				return nil, e
			}
		}

		return nil, err
	}

	response := newResponse(resp)

	err = CheckResponse(resp)
	if err != nil {
		defer resp.Body.Close()
		// Special case for AcceptedErrors. If an AcceptedError
		// has been encountered, the response's payload will be
		// added to the AcceptedError and returned.
		//
		// Issue #1022
		aerr, ok := err.(*AcceptedError)
		if ok {
			b, readErr := ioutil.ReadAll(resp.Body)
			if readErr != nil {
				return response, readErr
			}

			aerr.Raw = b
			err = aerr
		}
	}
	return response, err
}

// Do sends an API request and returns the API response. The API response is
// JSON decoded and stored in the value pointed to by v, or returned as an
// error if an API error has occurred. If v implements the io.Writer interface,
// the raw response body will be written to v, without attempting to first
// decode it. If v is nil, and no error hapens, the response is returned as is.
// If rate limit is exceeded and reset time is in the future, Do returns
// *RateLimitError immediately without making a network API call.
//
// The provided ctx must be non-nil, if it is nil an error is returned. If it
// is canceled or times out, ctx.Err() will be returned.
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*Response, error) {
	resp, err := c.BareDo(ctx, req)
	if err != nil {
		return resp, err
	}
	defer resp.Body.Close()

	switch v := v.(type) {
	case nil:
	case io.Writer:
		_, err = io.Copy(v, resp.Body)
	default:
		decErr := json.NewDecoder(resp.Body).Decode(v)
		if decErr == io.EOF {
			decErr = nil // ignore EOF errors caused by empty response body
		}
		if decErr != nil {
			err = decErr
		}
	}
	return resp, err
}

/*
An ErrorResponse reports one or more errors caused by an API request.

GitHub API docs: https://docs.github.com/en/free-pro-team@latest/rest/reference/#client-errors
*/
type ErrorResponse struct {
	Response *http.Response // HTTP response that caused this error
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d",
		r.Response.Request.Method, sanitizeURL(r.Response.Request.URL),
		r.Response.StatusCode)
}

// AcceptedError occurs when GitHub returns 202 Accepted response with an
// empty body, which means a job was scheduled on the GitHub side to process
// the information needed and cache it.
// Technically, 202 Accepted is not a real error, it's just used to
// indicate that results are not ready yet, but should be available soon.
// The request can be repeated after some time.
type AcceptedError struct {
	// Raw contains the response body.
	Raw []byte
}

func (*AcceptedError) Error() string {
	return "job scheduled on GitHub side; try again later"
}

// Is returns whether the provided error equals this error.
func (ae *AcceptedError) Is(target error) bool {
	v, ok := target.(*AcceptedError)
	if !ok {
		return false
	}
	return bytes.Equal(ae.Raw, v.Raw)
}

// sanitizeURL redacts the client_secret parameter from the URL which may be
// exposed to the user.
func sanitizeURL(uri *url.URL) *url.URL {
	if uri == nil {
		return nil
	}
	params := uri.Query()
	if len(params.Get("client_secret")) > 0 {
		params.Set("client_secret", "REDACTED")
		uri.RawQuery = params.Encode()
	}
	return uri
}

/*
An Error reports more details on an individual error in an ErrorResponse.
These are the possible validation error codes:

    missing:
        resource does not exist
    missing_field:
        a required field on a resource has not been set
    invalid:
        the formatting of a field is invalid
    already_exists:
        another resource has the same valid as this field
    custom:
        some resources return this (e.g. github.User.CreateKey()), additional
        information is set in the Message field of the Error

GitHub error responses structure are often undocumented and inconsistent.
Sometimes error is just a simple string (Issue #540).
In such cases, Message represents an error message as a workaround.

GitHub API docs: https://docs.github.com/en/free-pro-team@latest/rest/reference/#client-errors
*/
type Error struct {
	Resource string `json:"resource"` // resource on which the error occurred
	Field    string `json:"field"`    // field on which the error occurred
	Code     string `json:"code"`     // validation error code
	Message  string `json:"message"`  // Message describing the error. Errors with Code == "custom" will always have this set.
}

func (e *Error) Error() string {
	return fmt.Sprintf("%v error caused by %v field on %v resource",
		e.Code, e.Field, e.Resource)
}

func (e *Error) UnmarshalJSON(data []byte) error {
	type aliasError Error // avoid infinite recursion by using type alias.
	if err := json.Unmarshal(data, (*aliasError)(e)); err != nil {
		return json.Unmarshal(data, &e.Message) // data can be json string.
	}
	return nil
}

// CheckResponse checks the API response for errors, and returns them if
// present. A response is considered an error if it has a status code outside
// the 200 range or equal to 202 Accepted.
// API error responses are expected to have response
// body, and a JSON response body that maps to ErrorResponse.
//
// The error type will be *RateLimitError for rate limit exceeded errors,
// *AcceptedError for 202 Accepted status codes,
// and *TwoFactorAuthError for two-factor authentication errors.
func CheckResponse(r *http.Response) error {
	if r.StatusCode == http.StatusAccepted {
		return &AcceptedError{}
	}
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}

	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		err := json.Unmarshal(data, errorResponse)
		if err != nil {
			return err
		}
	}
	// Re-populate error response body because GitHub error responses are often
	// undocumented and inconsistent.
	// Issue #1136, #540.
	r.Body = ioutil.NopCloser(bytes.NewBuffer(data))
	return errorResponse
}
