package gthttp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/throttled/throttled"
	"github.com/throttled/throttled/store/memstore"
	"go-jw-toggl/jwtoggl/tools"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	DefaultAuthPassword       = "api_token"
	DefaultMaxRetries         = 5
	DefaultGzipEnabled        = false
	DefaultUrl                = "https://api.track.toggl.com/api/v9"
	DefaultVersion            = "v9"
	defaultBucket             = "toggl"
	DefaultRateLimitPerSecond = 1
)

// TogglHttpClient is a Toggl REST client. Created by calling NewClient.
type TogglHttpClient struct {
	client      *http.Client // net/http Client to use for requests
	version     string       // v8
	Url         string       // set of URLs passed initially to the client
	errorLog    Logger       // error log for critical messages
	infoLog     Logger       // information log for e.g. response times
	traceLog    Logger       // trace log for debugging
	password    string       // password for HTTP Basic Auth
	maxRetries  uint
	gzipEnabled bool // gzip compression enabled or disabled (default)
	rateLimiter *throttled.GCRARateLimiter
	perSec      int
	cookie      *http.Cookie
	token       string
}

type TogglError struct {
	Code       int
	Status     string
	Msg        string
	RetryAfter time.Duration
}

type TogglResponse struct {
	Data *json.RawMessage `json:"data"`
}

func (e *TogglError) Error() string {
	return fmt.Sprintf("%s\t%s\n", e.Status, e.Msg)
}

// ClientOptionFunc is a function that configures a Client.
// It is used in NewClient.
type ClientOptionFunc func(*TogglHttpClient) error

// NewClient returns a new TogglHttpClient. An error is also returned when some configuration option is invalid
//
//	tc,err := gtoggl.NewClient("token")
func NewClient(key string, options ...ClientOptionFunc) (*TogglHttpClient, error) {
	// Set up the client

	c := &TogglHttpClient{
		client:      http.DefaultClient,
		maxRetries:  DefaultMaxRetries,
		Url:         DefaultUrl,
		version:     DefaultVersion,
		gzipEnabled: DefaultGzipEnabled,
		password:    DefaultAuthPassword,
		errorLog:    defaultLogger,
		infoLog:     defaultLogger,
		traceLog:    defaultLogger,
		token:       key,
	}

	err := SetRateLimit(DefaultRateLimitPerSecond)(c)
	if err != nil {
		return nil, err
	}

	// Run the options on it
	for _, option := range options {
		if err := option(c); err != nil {
			return nil, err
		}
	}
	c.infoLog.Printf("Logging in with token: %s\n", key)

	if len(key) < 1 {
		c.errorLog.Printf("%s\n", "valid token required")
		return nil, errors.New("token required")
	}

	return c, nil
}

// SetHttpClient can be used to specify the http.Client to use when making
// HTTP requests to Toggl
func SetHttpClient(httpClient *http.Client) ClientOptionFunc {
	return func(c *TogglHttpClient) error {
		if httpClient != nil {
			c.client = httpClient
		} else {
			c.client = http.DefaultClient
		}
		return nil
	}
}

// SetURL defines the base URL. See DefaultUrl
//
//goland:noinspection GoUnusedExportedFunction
func SetURL(url string) ClientOptionFunc {
	return func(c *TogglHttpClient) error {
		c.Url = url
		return nil
	}
}

// SetRateLimit Set custom rate limit per second
func SetRateLimit(perSec int) ClientOptionFunc {
	return func(c *TogglHttpClient) error {
		store, err := memstore.New(65536)
		if err != nil {
			return err
		}
		quota := throttled.RateQuota{
			MaxRate:  throttled.PerSec(perSec),
			MaxBurst: 1,
		}
		c.rateLimiter, err = throttled.NewGCRARateLimiter(store, quota)
		if err != nil {
			return err
		}
		c.perSec = perSec
		return nil
	}
}

// SetTraceLogger sets a logger to print HTTP requests
//
//goland:noinspection GoUnusedExportedFunction
func SetTraceLogger(l Logger) ClientOptionFunc {
	return func(c *TogglHttpClient) error {
		c.traceLog = l
		return nil
	}
}

// SetErrorLogger sets a logger to handle error messages
//
//goland:noinspection GoUnusedExportedFunction
func SetErrorLogger(l Logger) ClientOptionFunc {
	return func(c *TogglHttpClient) error {
		c.errorLog = l
		return nil
	}
}

// SetInfoLogger sets a logger to handle info messages
//
//goland:noinspection GoUnusedExportedFunction
func SetInfoLogger(l Logger) ClientOptionFunc {
	return func(c *TogglHttpClient) error {
		c.infoLog = l
		return nil
	}
}

type nullLogger struct{}

//goland:noinspection GoUnusedParameter
func (l *nullLogger) Printf(format string, v ...interface{}) {
}

var defaultLogger = &nullLogger{}

func requestWithLimit(c *TogglHttpClient, method, endpoint string, b interface{}, attempt int) (*json.RawMessage, error) {
	c.infoLog.Printf("Request attempt %d for %s %s\n", attempt, method, endpoint)
	if attempt > DefaultMaxRetries {
		return nil, errors.New("Max Retries exceeded: " + strconv.FormatInt(DefaultMaxRetries, 10))
	}
	var body []byte
	var err error

	limited, reason, err := c.rateLimiter.RateLimit(defaultBucket, 1)
	if err != nil {
		return nil, err
	}

	if limited {
		c.traceLog.Printf("Hit rate limit. Sleeping for %f ms.\n", float64(reason.RetryAfter)/1000000)
		time.Sleep(reason.RetryAfter)
		return requestWithLimit(c, method, endpoint, b, attempt+1)
	}

	if b != nil {
		if body, err = json.Marshal(b); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.token, DefaultAuthPassword)
	c.dumpRequest(req)
	resp, err := c.client.Do(req)
	if err != nil {
		c.resetHttpClient("req", err)
		return nil, err
	}
	c.dumpResponse(resp)
	defer tools.MustClose(resp.Body)

	js, err := io.ReadAll(resp.Body)
	if err != nil {
		c.resetHttpClient("body", err)
		return nil, err
	}
	if resp.StatusCode == 429 {
		c.errorLog.Printf("Hit (429) rate limit. Sleeping for %d ms.\n", attempt*1000)
		time.Sleep(time.Millisecond * time.Duration(attempt*1000))
		return requestWithLimit(c, method, endpoint, b, attempt+1)
	}
	if resp.StatusCode == 404 {
		return nil, nil
	}
	if resp.StatusCode >= 400 {
		retryAfter := time.Duration(0)
		if resp.StatusCode == http.StatusPaymentRequired {
			if seconds, err := strconv.Atoi(resp.Header.Get("X-Toggl-Quota-Resets-In")); err == nil && seconds > 0 {
				retryAfter = time.Duration(seconds) * time.Second
			}
		}
		if resp.StatusCode == http.StatusPaymentRequired {
			return nil, &TogglError{Code: resp.StatusCode, Status: resp.Status, Msg: string(js), RetryAfter: retryAfter}
		}
		c.resetHttpClient("status", fmt.Errorf("response status %d", resp.StatusCode))
		return nil, &TogglError{Code: resp.StatusCode, Status: resp.Status, Msg: string(js)}
	}
	var raw json.RawMessage
	if json.Unmarshal(js, &raw) != nil {
		return nil, err
	}
	return &raw, err
}

func request(c *TogglHttpClient, method, endpoint string, b interface{}) (*json.RawMessage, error) {
	return requestWithLimit(c, method, endpoint, b, 1)
}

// PostRequest makes POST request
func (c *TogglHttpClient) PostRequest(endpoint string, body interface{}) (*json.RawMessage, error) {
	return request(c, "POST", endpoint, body)
}

// DeleteRequest makes DELETE request
func (c *TogglHttpClient) DeleteRequest(endpoint string, body interface{}) (*json.RawMessage, error) {
	return request(c, "DELETE", endpoint, body)
}

// PutRequest makes PUT request
func (c *TogglHttpClient) PutRequest(endpoint string, body interface{}) (*json.RawMessage, error) {
	return request(c, "PUT", endpoint, body)
}

// GetRequest makes GET request
func (c *TogglHttpClient) GetRequest(endpoint string) (*json.RawMessage, error) {
	return request(c, "GET", endpoint, nil)
}

func (c *TogglHttpClient) resetHttpClient(reason string, err error) {
	log.Printf("Resetting client (%s): %+v", reason, err)
	c.client.CloseIdleConnections()
	c.cookie = nil
}
