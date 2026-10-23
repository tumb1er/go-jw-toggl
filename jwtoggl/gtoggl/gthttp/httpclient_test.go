package gthttp

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestDefaultRateLimitMatchesTogglSafeRate(t *testing.T) {
	if DefaultRateLimitPerSecond != 1 {
		t.Fatalf("DefaultRateLimitPerSecond = %d, want 1", DefaultRateLimitPerSecond)
	}
}

func TestRequestReturnsQuotaResetForHTTP402(t *testing.T) {
	client, err := NewClient("offline-test-token", SetHttpClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Host != "toggl.test" {
				t.Fatalf("request host = %q, want toggl.test", req.URL.Host)
			}
			return &http.Response{
				StatusCode: http.StatusPaymentRequired,
				Status:     "402 Payment Required",
				Header:     http.Header{"X-Toggl-Quota-Resets-In": []string{"37"}},
				Body:       ioutil.NopCloser(strings.NewReader(`{"message":"quota exceeded"}`)),
				Request:    req,
			}, nil
		}),
	}))
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.GetRequest("https://toggl.test/quota")
	var togglErr *TogglError
	if !errors.As(err, &togglErr) {
		t.Fatalf("error = %T %v, want *TogglError", err, err)
	}
	if togglErr.RetryAfter != 37*time.Second {
		t.Fatalf("RetryAfter = %s, want 37s", togglErr.RetryAfter)
	}
}
