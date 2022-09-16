package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/xefino/goutils/collections"
	"github.com/xefino/goutils/utils"
)

// Defines codes that should result in a retry when encountered
var retryCodes = []int{http.StatusBadGateway, http.StatusRequestTimeout,
	http.StatusConflict, http.StatusTooManyRequests}

// WebClient defines an HTTP client that can be used to handle typical JSON responses from an API
type WebClient struct {
	client        *http.Client
	startInterval time.Duration
	endInterval   time.Duration
	maxElapsed    time.Duration
	retryCodes    []int
	errorHandler  func(*WebClient, []byte) string
	logger        *utils.Logger
}

// NewWebClient creates a new connection to an API
func NewWebClient(logger *utils.Logger, opts ...IWebClientOption) *WebClient {
	return WithClient(new(http.Client), logger, opts...)
}

// WithClient creates a new connection to the API with a given HTTP client
func WithClient(client *http.Client, logger *utils.Logger, opts ...IWebClientOption) *WebClient {

	// First, create the web client with our default values
	wc := WebClient{
		client:        client,
		startInterval: 500,
		endInterval:   60000,
		maxElapsed:    900000,
		retryCodes:    retryCodes,
		errorHandler:  nil,
		logger:        logger.ChangeFrame(3),
	}

	// Next, call each of our options to modify the client
	for _, opt := range opts {
		opt.Apply(&wc)
	}

	// Finally, return a pointer to the client
	return &wc
}

// GetData attempts to run an HTTP request against an endpoint and deserialize the respone into the object provided
func (client *WebClient) GetData(request *http.Request, obj interface{}) error {

	// First, attempt to get the data from the endpoint; return any error that occurs
	resp, err := client.DoRequest(request)
	if err != nil {
		return err
	}

	// Next, attempt to read the body of our response and close it when we're done
	// If this fails then return an error
	defer resp.Body.Close()
	body, err := client.GetBody(resp.Body)
	if err != nil {
		return err
	}

	// Finally, attempt to deserialize the body from JSON; if this fails then return an error
	if err := client.Deserialize(body, obj); err != nil {
		return err
	}

	return nil
}

// DoRequest attempts an HTTP request and returns the HTTP response
func (client *WebClient) DoRequest(request *http.Request) (*http.Response, error) {
	client.logger.Log("Requesting page from %s...", request.URL)

	// Attempt the request with an exponential backoff so that we can retry on failures
	var resp *http.Response
	err := backoff.Retry(func() error {
		var err error

		// If the request returns an error, the respone is nil or the response status code
		// indicates that the problem will not be resolved witha retry then embed the response
		// into an error and return it. If the response status code is not 2xx then retry
		if resp, err = client.client.Do(request); err != nil || resp == nil ||
			(!collections.Contains(retryCodes, resp.StatusCode) && resp.StatusCode >= 400) {
			if err != nil {
				return backoff.Permanent(err)
			} else {
				return backoff.Permanent(fmt.Errorf("unrecoverable error occurred"))
			}
		} else if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			client.logger.Log("Request to %s failed with error code %d. Retrying...",
				request.URL.String(), resp.StatusCode)
			return fmt.Errorf("maximum retry count exceeded")
		}

		return nil
	}, client.createExponentialBackoff())

	// If the request returned an error then embed it into a respone and return it
	if err != nil {
		return resp, client.FromHTTPResponse(err, resp)
	}

	return resp, nil
}

// GetBody reads the body from an HTTP response
func (client *WebClient) GetBody(reader io.ReadCloser) ([]byte, error) {
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, client.NewClientError(err, "Error reading response body")
	}

	return body, nil
}

// Deserialize extracts the respone body JSON into the object provided
func (client *WebClient) Deserialize(body []byte, obj interface{}) error {

	// Remove the BOM from the response
	body = bytes.TrimPrefix(body, []byte("\xef\xbb\xbf"))

	// Attempt to unmarshal the object from the body; encapsulate any error
	// in our Error object and return it if this fails
	if err := json.Unmarshal(body, obj); err != nil {
		return client.NewClientError(err, "Failed to unmarsahl JSON response body")
	}

	return nil
}

// Helper function that can be used to create an exponential backoff
// timer from values stored on the client
func (client *WebClient) createExponentialBackoff() *backoff.ExponentialBackOff {

	// Create the timer with values from the requester and some values that
	// are standard to all exponential backoff timers from the backoff library
	timer := backoff.NewExponentialBackOff()
	timer.InitialInterval = client.startInterval * time.Millisecond
	timer.MaxInterval = client.endInterval * time.Millisecond
	timer.MaxElapsedTime = client.maxElapsed * time.Millisecond
	timer.RandomizationFactor = backoff.DefaultRandomizationFactor
	timer.Multiplier = backoff.DefaultMultiplier
	timer.Clock = backoff.SystemClock

	// Reset the timer and return it
	timer.Reset()
	return timer
}
