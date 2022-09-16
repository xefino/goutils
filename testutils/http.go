package testutils

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/xefino/goutils/strings"
)

// GenerateResponse creates an HTTP response we can use in our testing
func GenerateResponse(request *http.Request, code int, item string) *http.Response {

	// If we have data then enclose it in a ReadCloser, otherwise set it to NoBody
	var body io.ReadCloser
	if strings.IsEmpty(item) {
		body = http.NoBody
	} else {
		body = ioutil.NopCloser(bytes.NewBuffer([]byte(item)))
	}

	// Create the response and return it
	return &http.Response{
		StatusCode: code,
		Body:       body,
		Request:    request,
		Header:     make(http.Header),
	}
}

// VerifyRequest verifies the details of an HTTP request
func VerifyRequest(req *http.Request, method string, uri string) {
	defer GinkgoRecover()
	Expect(req.URL.String()).Should(Equal(uri))
	Expect(req.Method).Should(Equal(method))
	Expect(req.Header.Get("Authorization")).Should(Equal("Bearer FAKE_KEY"))
}

// VerifyAndGenerateResponse verifies the HTTP request and generates an HTTP response
func VerifyAndGenerateResponse(method string, uri string, code int,
	data string) func(*http.Request) *http.Response {
	return func(req *http.Request) *http.Response {
		VerifyRequest(req, method, uri)
		return GenerateResponse(req, code, data)
	}
}

// RoundTripFunc allows us to encapsulate the functionality necessary to mock
// HTTP requests made from an HTTP client so we can test our request functions
type RoundTripFunc struct {
	Functions  []func(*http.Request) *http.Response
	Index      int
	ShouldFail bool
}

// RoundTrip runs a single HTTP request
func (f *RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {

	// First, get the response with the function and prepare for the next call
	response := f.Functions[f.Index](req)
	f.Index++

	// If we want to return an error then do so here
	if f.ShouldFail {
		return nil, fmt.Errorf("RoundTrip failed")
	}

	return response, nil
}

// NewTestClient returns an HTTP client with the transport replaced to avoid
// make real HTTP requests against the endpionts sent to it
func NewTestClient(shouldFail bool, fns ...func(*http.Request) *http.Response) *http.Client {
	return &http.Client{
		Transport: &RoundTripFunc{
			Functions:  fns,
			ShouldFail: shouldFail,
		},
	}
}

// ErrorReader is a mock type we'll use in place of a reader so we
// can test out errors when Read is called
type ErrorReader int

// Read allows for the Reader interface to be implemented and returns an error
func (ErrorReader) Read([]byte) (int, error) {
	return 0, fmt.Errorf("Read failed")
}
