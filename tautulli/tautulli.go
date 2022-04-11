package tautulli

import (
	"context"
	"fmt"
	"github.com/google/go-querystring/query"
	"net/http"
	"net/url"
	"strings"
)

const (
	tautulliAPIPath = "/api/v2"
	userAgent       = "go-tautulli"
)

//go:generate mockery --name=Client --case=snake
// A Client manages communication with the Tautulli API.
type Client struct {
	// HTTP client used to communicate with the API.
	client *http.Client
	// Base URL for API requests. Defaults to the public GitHub API, but can be
	// set to a domain endpoint to use with GitHub Enterprise. BaseURL should
	// always be specified with a trailing slash.
	BaseURL *url.URL
	// API Key for Tautulli
	APIKey string
	// User agent used when communicating with the GitHub API.
	UserAgent string

	common service // Reuse a single struct instead of allocating one for each service on the heap.

	Notifications *NotificationsService
}

type service struct {
	client *Client
}

// CommonParameters are the parameters that are common to all requests.
type CommonParameters struct {
	APIKey   string `url:"apikey"`   //
	Command  string `url:"cmd"`      //
	OutType  string `url:"out_type"` //
	Callback string `url:"callback"` //
	Debug    int    `url:"debug"`    //
}

// Client returns the http.Client used by this Tautulli client.
func (c *Client) Client() *http.Client {
	clientCopy := *c.client
	return &clientCopy
}

// NewClient returns a new Tautulli API client. If a nil httpClient is
// provided, a new http.Client will be used.
func NewClient(httpClient *http.Client, baseURL *url.URL, apiKey string) *Client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	c := &Client{client: httpClient, BaseURL: baseURL, APIKey: apiKey, UserAgent: userAgent}
	c.common.client = c
	c.Notifications = (*NotificationsService)(&c.common)
	return c
}

// NewCommand creates an API request. A relative URL can be provided in urlStr,
// in which case it is resolved relative to the BaseURL of the Client.
// Relative URLs should always be specified without a preceding slash. If
// specified, the value pointed to by body is JSON encoded and included as the
// request body.
func (c *Client) NewCommand(command string, params interface{}) (*http.Request, error) {
	if !strings.HasSuffix(c.BaseURL.Path, "/") {
		return nil, fmt.Errorf("BaseURL must have a trailing slash, but %q does not", c.BaseURL)
	}

	a, _ := query.Values(CommonParameters{
		APIKey:   c.common.client.APIKey,
		Command:  command,
		OutType:  "json",
		Callback: "pong",
		Debug:    1,
	})
	v, _ := query.Values(params)
	u, err := c.BaseURL.Parse(tautulliAPIPath + "?" + a.Encode() + "&" + v.Encode())
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	return req, nil
}

// Do sends an API request and returns the API response.
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	switch resp.StatusCode {
	case http.StatusOK:
		return resp, nil
	case http.StatusBadRequest:
		return resp, fmt.Errorf("bad request")
	case http.StatusUnauthorized:
		return resp, fmt.Errorf("unauthorized")
	case http.StatusNotFound:
		return resp, fmt.Errorf("not found")
	default:
		return resp, fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}
}
