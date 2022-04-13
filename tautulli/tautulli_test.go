package tautulli

import (
	"context"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

const (
	// baseURLPath is a non-empty Client.BaseURL path to use during tests,
	// to ensure relative URLs are used for all endpoints.
	baseURLPath = "/api/v2"
)

// setup sets up a test HTTP server along with a github.Client that is
// configured to talk to that test server. Tests should register handlers on
// mux which provide mock responses for the API method being tested.
func setup() (client *Client, mux *http.ServeMux, serverURL string, teardown func()) {
	// mux is the HTTP request multiplexer used with the test server.
	mux = http.NewServeMux()

	// We want to ensure that tests catch mistakes where the endpoint URL is
	// specified as absolute rather than relative. It only makes a difference
	// when there's a non-empty base URL path. So, use that. See issue #752.
	apiHandler := http.NewServeMux()
	apiHandler.Handle(baseURLPath+"/", http.StripPrefix(baseURLPath, mux))
	apiHandler.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(os.Stderr, "FAIL: Client.BaseURL path prefix is not preserved in the request URL:")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "\t"+req.URL.String())
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "\tDid you accidentally use an absolute endpoint URL rather than relative?")
		fmt.Fprintln(os.Stderr, "\tSee https://github.com/google/go-github/issues/752 for information.")
		http.Error(w, "Client.BaseURL path prefix is not preserved in the request URL.", http.StatusInternalServerError)
	})

	// server is a test HTTP server used to provide mock API responses.
	server := httptest.NewServer(apiHandler)

	baseURL, _ := url.Parse(server.URL)

	// client is the Tautulli client being tested and is
	// configured to use test server.
	client = NewClient(nil, baseURL, "test", &ClientOptions{
		Debug: true,
	})

	return client, mux, server.URL, server.Close
}

func testMethod(t *testing.T, r *http.Request, want string) {
	t.Helper()
	if got := r.Method; got != want {
		t.Errorf("Request method: %v, want %v", got, want)
	}
}

type values map[string]string

func testFormValues(t *testing.T, r *http.Request, values values) {
	t.Helper()
	want := url.Values{}
	for k, v := range values {
		want.Set(k, v)
	}

	_ = r.ParseForm()
	if got := r.Form; !cmp.Equal(got, want) {
		t.Errorf("Request parameters: %v, want %v", got, want)
	}
}

func TestTautulli_ErrorResponse(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc(baseURLPath, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"message": "Bad request."}`)
	})

	ctx := context.Background()
	params := &NotifyParameters{
		NotifierID: 0,
		Subject:    "",
	}
	_, err := client.Notifications.Notify(ctx, params)
	if err == nil {
		t.Errorf("Notifications.Notify returned nil error")
	}
}

func TestTautulli_TestNewClient(t *testing.T) {
	c := NewClient(nil, nil, "", nil)

	if got, want := c.UserAgent, userAgent; got != want {
		t.Errorf("NewClient UserAgent is %v, want %v", got, want)
	}

	c2 := NewClient(nil, nil, "", nil)
	if c.client == c2.client {
		t.Error("NewClient returned same http.Clients, but they should differ")
	}
}

func TestTautulli_TestClient(t *testing.T) {
	c := NewClient(nil, nil, "", nil)
	c2 := c.Client()
	if c.client == c2 {
		t.Error("Client returned same http.Client, but should be different")
	}
}
