package http

import (
	"fmt"
	"net/http"

	"github.com/xefino/goutils/strings"
	"github.com/xefino/goutils/utils"
)

// Error describes an error returned by the Polygon client
type Error struct {
	*utils.GError
	StatusCode int
}

// NewClientError creates a new client error from the original error,
// an erorr message and associated format arguments
func (client *WebClient) NewClientError(original error, message string, args ...interface{}) *Error {
	return &Error{GError: client.logger.Error(original, message, args...)}
}

// FromHTTPResponse creates a new client error from an HTTP response,
// and an inner error
func (client *WebClient) FromHTTPResponse(err error, resp *http.Response) *Error {

	// If the response isn't nil then create a message and inject it,
	// and the associated status code into the error and return it
	if resp != nil {
		defer resp.Body.Close()

		// First, if the error is the result of a failure on the Polygon
		// API, then we'll want to extract that message. It could either
		// be in the error field or in the message field
		var inner string
		if data, bErr := client.GetBody(resp.Body); bErr == nil && client.errorHandler != nil {
			inner = client.errorHandler(client, data)
		}

		// Next, if we managed to extract the inner message then add an
		// identifier to it so the reader can understand it
		if !strings.IsEmpty(inner) {
			inner = ", Inner Error: " + inner
		}

		// Finally, generate the whole message and generate an error from it
		message := fmt.Sprintf("API request to %s failed, %s response returned%s",
			resp.Request.URL.String(), http.StatusText(resp.StatusCode), inner)
		return &Error{
			GError:     client.logger.Error(err, message),
			StatusCode: resp.StatusCode,
		}
	}

	// Otherwise, create a standard error message and return it
	return &Error{GError: client.logger.Error(err, "API request failed; no response received")}
}
