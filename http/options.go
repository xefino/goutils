package http

import "time"

// IWebClientOption defines the functionality that will allow the behavior of a
// WebClient to be modified at construction
type IWebClientOption interface {
	Apply(*WebClient)
}

// WithErrorHandler allows the user to set the error handler function that is called
// when an API request returns a bad response
type WithErrorHandler func(*WebClient, []byte) string

// Apply modifies the WebClient so that it has the error handler defined by this object
func (w WithErrorHandler) Apply(client *WebClient) {
	client.errorHandler = w
}

// WithBackoffStart allows the user to set the starting time to use when backing off from
// an API error that should be retried
type WithBackoffStart time.Duration

// Apply modifies the WebClient so that it has the start interval defined by this object
func (w WithBackoffStart) Apply(client *WebClient) {
	client.startInterval = time.Duration(w)
}

// WithBackoffEnd allows the user to set the ending time to use when backing off from an
// API error that should be retried
type WithBackoffEnd time.Duration

// Apply modifies the WebClient so that it has the end interval defined by this object
func (w WithBackoffEnd) Apply(client *WebClient) {
	client.endInterval = time.Duration(w)
}

// WithBackoffMaxElapsed allows the user to set the maximum time that should be allowed when
// the API returns an error that should be retried
type WithBackoffMaxElapsed time.Duration

// Apply modifies the WebClient so that it has the maximum interval defined by this object
func (w WithBackoffMaxElapsed) Apply(client *WebClient) {
	client.maxElapsed = time.Duration(w)
}

// WithRetryCodes allows the user to define the HTTP status codes that would trigger a retry of
// the API endpoint rather than generating an error
type WithRetryCodes []int

// Apply modifies the WebClient so that it has the retry codes defined by this object
func (w WithRetryCodes) Apply(client *WebClient) {
	client.retryCodes = w
}
