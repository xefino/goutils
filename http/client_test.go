package http

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/xefino/goutils/testutils"
	"github.com/xefino/goutils/utils"
)

// Create a new test runner we'll use to test all the
// modules in the http package
func TestHttp(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HTTP Suite")
}

var _ = Describe("WebClient Tests", func() {

	// Tests that the constructor works as expected
	It("Constructor - Works", func() {

		// First, create our test logger and ensure that its
		// output is discarded
		logger := utils.NewLogger("testd", "test")
		logger.Discard()

		// Next, inject the logger and API key into our client
		client := NewWebClient(logger)

		// Finally, verify that the client was created
		Expect(client.startInterval).Should(Equal(time.Duration(500)))
		Expect(client.endInterval).Should(Equal(time.Duration(60000)))
		Expect(client.maxElapsed).Should(Equal(time.Duration(900000)))
		Expect(client.client).ShouldNot(BeNil())
		Expect(client.logger).ShouldNot(BeNil())
		Expect(client.logger).ShouldNot(Equal(logger))
	})

	// Test that doRequest returns an error if the HTTP client returns an error
	It("DoRequest - HTTP Error Returned - Error", func() {

		// Create the test client
		httpClient := testutils.NewTestClient(true, testutils.VerifyAndGenerateResponse(http.MethodGet,
			"test.url/fails", http.StatusBadRequest, ""))

		// Create the web client from the test client
		client := generateClient(httpClient)

		// Create the HTTP request
		request, _ := http.NewRequest(http.MethodGet, "test.url/fails", http.NoBody)
		request.Header.Add("Authorization", "Bearer FAKE_KEY")

		// Attempt to send the request; this should fail
		resp, err := client.DoRequest(request)
		actual := err.(*Error)

		// Verify the failure
		Expect(resp).Should(BeNil())
		Expect(actual).Should(HaveOccurred())
		Expect(actual.Class).Should(Equal("WebClient"))
		Expect(actual.Environment).Should(Equal("test"))
		Expect(actual.File).Should(Equal("/goutils/http/client.go"))
		Expect(actual.Function).Should(Equal("DoRequest"))
		Expect(actual.GeneratedAt).ShouldNot(BeNil())
		Expect(actual.Inner).Should(HaveOccurred())
		Expect(actual.Inner.Error()).Should(Equal("Get \"test.url/fails\": RoundTrip failed"))
		Expect(actual.LineNumber).Should(Equal(115))
		Expect(actual.Message).Should(Equal("API request failed; no response received"))
		Expect(actual.Package).Should(Equal("http"))
		Expect(actual.StatusCode).Should(BeZero())
		Expect(actual.Error()).Should(HaveSuffix("[test] http.WebClient.DoRequest (/goutils/http/client.go 115): " +
			"API request failed; no response received, Inner:\n\tGet \"test.url/fails\": RoundTrip failed."))
	})

	// Test that doRequest returns an error if the HTTP response code is less than 200
	It("DoRequest - Response code < 200 - Error", func() {

		// First, create the test client with the requests we expect
		httpClient := testutils.NewTestClient(false,
			testutils.VerifyAndGenerateResponse(http.MethodGet, "test.url/fails", http.StatusContinue, ""),
			testutils.VerifyAndGenerateResponse(http.MethodGet, "test.url/fails", http.StatusContinue, ""),
			testutils.VerifyAndGenerateResponse(http.MethodGet, "test.url/fails", http.StatusContinue, ""),
			testutils.VerifyAndGenerateResponse(http.MethodGet, "test.url/fails", http.StatusContinue, ""),
			testutils.VerifyAndGenerateResponse(http.MethodGet, "test.url/fails", http.StatusContinue, ""),
			testutils.VerifyAndGenerateResponse(http.MethodGet, "test.url/fails", http.StatusContinue, ""))

		// Next, create the web client from the test client
		client := generateClient(httpClient)

		// Now, create the HTTP request
		request, _ := http.NewRequest(http.MethodGet, "test.url/fails", http.NoBody)
		request.Header.Add("Authorization", "Bearer FAKE_KEY")

		// Finally, attempt to send the request; this should fail
		resp, err := client.DoRequest(request)
		actual := err.(*Error)

		// Verify the failure
		Expect(resp).ShouldNot(BeNil())
		Expect(resp.StatusCode).Should(Equal(http.StatusContinue))
		Expect(actual).Should(HaveOccurred())
		Expect(actual.Class).Should(Equal("WebClient"))
		Expect(actual.Environment).Should(Equal("test"))
		Expect(actual.File).Should(Equal("/goutils/http/client.go"))
		Expect(actual.Function).Should(Equal("DoRequest"))
		Expect(actual.GeneratedAt).ShouldNot(BeNil())
		Expect(actual.Inner).Should(HaveOccurred())
		Expect(actual.Inner.Error()).Should(Equal("maximum retry count exceeded"))
		Expect(actual.LineNumber).Should(Equal(115))
		Expect(actual.Message).Should(Equal("API request to test.url/fails failed, " +
			"Continue response returned, Inner Error: TEST ERROR"))
		Expect(actual.Package).Should(Equal("http"))
		Expect(actual.StatusCode).Should(Equal(100))
		Expect(actual.Error()).Should(HaveSuffix("[test] http.WebClient.DoRequest (/goutils/http/client.go 115): " +
			"API request to test.url/fails failed, Continue response returned, Inner Error: TEST ERROR, " +
			"Inner:\n\tmaximum retry count exceeded."))
	})

	// Test that doRequest returns an error if the HTTP response code is greater than 299
	It("DoRequest - Response code > 299 - Error", func() {

		// First, create the test client with the requests we expect
		httpClient := testutils.NewTestClient(false,
			testutils.VerifyAndGenerateResponse(http.MethodGet, "test.url/fails", http.StatusMultipleChoices, ""),
			testutils.VerifyAndGenerateResponse(http.MethodGet, "test.url/fails", http.StatusMultipleChoices, ""),
			testutils.VerifyAndGenerateResponse(http.MethodGet, "test.url/fails", http.StatusMultipleChoices, ""),
			testutils.VerifyAndGenerateResponse(http.MethodGet, "test.url/fails", http.StatusMultipleChoices, ""),
			testutils.VerifyAndGenerateResponse(http.MethodGet, "test.url/fails", http.StatusMultipleChoices, ""),
			testutils.VerifyAndGenerateResponse(http.MethodGet, "test.url/fails", http.StatusMultipleChoices, ""))

		// Next, create the web client from the test client
		client := generateClient(httpClient)

		// Now, create the HTTP request
		request, _ := http.NewRequest(http.MethodGet, "test.url/fails", http.NoBody)
		request.Header.Add("Authorization", "Bearer FAKE_KEY")

		// Finally, attempt to send the request; this should fail
		resp, err := client.DoRequest(request)
		actual := err.(*Error)

		// Verify the failure
		Expect(resp).ShouldNot(BeNil())
		Expect(resp.StatusCode).Should(Equal(http.StatusMultipleChoices))
		Expect(actual).Should(HaveOccurred())
		Expect(actual.Class).Should(Equal("WebClient"))
		Expect(actual.Environment).Should(Equal("test"))
		Expect(actual.File).Should(Equal("/goutils/http/client.go"))
		Expect(actual.Function).Should(Equal("DoRequest"))
		Expect(actual.GeneratedAt).ShouldNot(BeNil())
		Expect(actual.Inner).Should(HaveOccurred())
		Expect(actual.Inner.Error()).Should(Equal("maximum retry count exceeded"))
		Expect(actual.LineNumber).Should(Equal(115))
		Expect(actual.Message).Should(Equal("API request to test.url/fails failed, " +
			"Multiple Choices response returned, Inner Error: TEST ERROR"))
		Expect(actual.Package).Should(Equal("http"))
		Expect(actual.StatusCode).Should(Equal(300))
		Expect(actual.Error()).Should(HaveSuffix("[test] http.WebClient.DoRequest (/goutils/http/client.go 115): " +
			"API request to test.url/fails failed, Multiple Choices response returned, Inner Error: TEST ERROR, " +
			"Inner:\n\tmaximum retry count exceeded."))
	})

	// Test that doRequest returns an error if the HTTP response code is greater than or
	// equal to 400 without attempting a retry
	It("DoRequest - Response code >= 400 - No retry, Error", func() {

		// First, create the test client with the requests we expect
		httpClient := testutils.NewTestClient(false,
			testutils.VerifyAndGenerateResponse(http.MethodGet, "test.url/fails", http.StatusBadRequest,
				"{\"status\":\"ERROR\",\"request_id\":\"dff53d9c74f4edff15348f523f2ee922\","+
					"\"error\":\"Failed to parse query parameters from URL: strconv.ParseBool: parsing \\\"derp\\\": invalid syntax\"}"))

		// Next, create the web client from the test client
		client := generateClient(httpClient)

		// Now, create the HTTP request
		request, _ := http.NewRequest(http.MethodGet, "test.url/fails", http.NoBody)
		request.Header.Add("Authorization", "Bearer FAKE_KEY")

		// Finally, attempt to send the request; this should fail
		resp, err := client.DoRequest(request)
		actual := err.(*Error)

		// Verify the failure
		Expect(resp).ShouldNot(BeNil())
		Expect(resp.StatusCode).Should(Equal(http.StatusBadRequest))
		Expect(actual).Should(HaveOccurred())
		Expect(actual.Class).Should(Equal("WebClient"))
		Expect(actual.Environment).Should(Equal("test"))
		Expect(actual.File).Should(Equal("/goutils/http/client.go"))
		Expect(actual.Function).Should(Equal("DoRequest"))
		Expect(actual.GeneratedAt).ShouldNot(BeNil())
		Expect(actual.Inner).Should(HaveOccurred())
		Expect(actual.Inner.Error()).Should(Equal("unrecoverable error occurred"))
		Expect(actual.LineNumber).Should(Equal(115))
		Expect(actual.Message).Should(Equal("API request to test.url/fails failed, " +
			"Bad Request response returned, Inner Error: TEST ERROR"))
		Expect(actual.Package).Should(Equal("http"))
		Expect(actual.StatusCode).Should(Equal(400))
		Expect(actual.Error()).Should(HaveSuffix("[test] http.WebClient.DoRequest (/goutils/http/client.go 115): " +
			"API request to test.url/fails failed, Bad Request response returned, Inner Error: TEST ERROR, " +
			"Inner:\n\tunrecoverable error occurred."))
	})

	// Test that GetBody fails if the ReadAll function returns an error
	It("GetBody - ReadAll Error Returned - Error", func() {

		// Create the web client with no underlying HTTP client
		client := generateClient(nil)

		// Attempt to get an invalid response body; this should fail
		body, err := client.GetBody(ioutil.NopCloser(testutils.ErrorReader(0)))
		actual := err.(*Error)

		// Verify failure
		Expect(body).Should(BeEmpty())
		Expect(actual).Should(HaveOccurred())
		Expect(actual.Class).Should(Equal("WebClient"))
		Expect(actual.Environment).Should(Equal("test"))
		Expect(actual.File).Should(Equal("/goutils/http/client.go"))
		Expect(actual.Function).Should(Equal("GetBody"))
		Expect(actual.GeneratedAt).ShouldNot(BeNil())
		Expect(actual.Inner).Should(HaveOccurred())
		Expect(actual.Inner.Error()).Should(Equal("Read failed"))
		Expect(actual.LineNumber).Should(Equal(125))
		Expect(actual.Message).Should(Equal("Error reading response body"))
		Expect(actual.Package).Should(Equal("http"))
		Expect(actual.StatusCode).Should(BeZero())
		Expect(actual.Error()).Should(HaveSuffix("[test] http.WebClient.GetBody (/goutils/http/client.go 125): " +
			"Error reading response body, Inner:\n\tRead failed."))
	})

	// Test that GetBody succeeds if the ReadAll function works
	It("GetBody - No Error - Body Returned", func() {

		// Create the web client with no underlying HTTP client
		client := generateClient(nil)

		// Attempt to get an invalid response body; this should fail
		body, err := client.GetBody(ioutil.NopCloser(bytes.NewBufferString("OK")))

		// Verify the success
		Expect(body).Should(Equal([]byte("OK")))
		Expect(err).ShouldNot(HaveOccurred())
	})

	// Test that Deserialize will return an error if the data canont
	// be deserialized to the object provided
	It("Deserialize - Unmarshal error - Error", func() {

		// First, create a test payload we'll try to deserialize
		data := []byte("{\"Key\":\"herp\",\"Value\":\"derp\"}")

		// Next, create a web client with mock data
		client := generateClient(nil)

		// Now, create a test type that we'll try to deserialize to
		value := struct {
			Key   string
			Value int
		}{}

		// Finally, attempt to deserialize the test data to this type; this should fail
		err := client.Deserialize(data, &value)
		actual := err.(*Error)

		// Verify the failure
		Expect(actual).Should(HaveOccurred())
		Expect(actual.Class).Should(Equal("WebClient"))
		Expect(actual.Environment).Should(Equal("test"))
		Expect(actual.File).Should(Equal("/goutils/http/client.go"))
		Expect(actual.Function).Should(Equal("Deserialize"))
		Expect(actual.GeneratedAt).ShouldNot(BeNil())
		Expect(actual.Inner).Should(HaveOccurred())
		Expect(actual.Inner.Error()).Should(Equal("json: cannot unmarshal string into Go struct field .Value of type int"))
		Expect(actual.LineNumber).Should(Equal(140))
		Expect(actual.Message).Should(Equal("Failed to unmarsahl JSON response body"))
		Expect(actual.Package).Should(Equal("http"))
		Expect(actual.StatusCode).Should(BeZero())
		Expect(actual.Error()).Should(HaveSuffix("[test] http.WebClient.Deserialize (/goutils/http/client.go 140): " +
			"Failed to unmarsahl JSON response body, Inner:\n\tjson: cannot unmarshal string into Go struct field " +
			".Value of type int."))
	})

	// Test that deserialize will successfully deserialize the data,
	// even if the data has a BOM prefixed
	It("Deserialize - Body contains BOM - Works", func() {

		// First, create a test payload we'll try to deserialize
		data := []byte("{\"Key\":\"herp\",\"Value\":\"derp\"}")

		// Next, create a web client with mock data
		client := generateClient(nil)

		// Now, add the BOM to the data and then attempt to
		// deserialize it; this should succeed
		var result test
		data = append([]byte("\xef\xbb\xbf"), data...)
		err := client.Deserialize(data, &result)

		// Finally, verify the data
		Expect(err).ShouldNot(HaveOccurred())
		Expect(result.Key).Should(Equal("herp"))
		Expect(result.Value).Should(Equal("derp"))
	})

	// Test that Deserialize will successfully deserialize the data,
	// even if the data doesn't have a BOM prefixed
	It("Deserialize - Body does not contain BOM - Works", func() {

		// First, create a test payload we'll try to deserialize
		data := []byte("{\"Key\":\"herp\",\"Value\":\"derp\"}")

		// Next, create a web client with mock data
		client := generateClient(nil)

		// Now, add the BOM to the data and then attempt to
		// deserialize it; this should succeed
		var result test
		err := client.Deserialize(data, &result)

		// Finally, verify the data
		Expect(err).ShouldNot(HaveOccurred())
		Expect(result.Key).Should(Equal("herp"))
		Expect(result.Value).Should(Equal("derp"))
	})

	// Test that GetData will fail to get the data if the HTTP request fails
	It("GetData - DoRequest fails - Error", func() {

		// First, create the test client
		httpClient := testutils.NewTestClient(true, testutils.VerifyAndGenerateResponse(http.MethodGet,
			"test.url/fails", http.StatusBadRequest, ""))

		// Next, create the web client from the test client
		client := generateClient(httpClient)

		// Now, create the HTTP request
		request, _ := http.NewRequest(http.MethodGet, "test.url/fails", http.NoBody)
		request.Header.Add("Authorization", "Bearer FAKE_KEY")

		// Finally, attempt to send the request; this should fail
		var value test
		err := client.GetData(request, &value)
		actual := err.(*Error)

		// Verify the failure
		Expect(actual).Should(HaveOccurred())
		Expect(actual.Class).Should(Equal("WebClient"))
		Expect(actual.Environment).Should(Equal("test"))
		Expect(actual.File).Should(Equal("/goutils/http/client.go"))
		Expect(actual.Function).Should(Equal("DoRequest"))
		Expect(actual.GeneratedAt).ShouldNot(BeNil())
		Expect(actual.Inner).Should(HaveOccurred())
		Expect(actual.Inner.Error()).Should(Equal("Get \"test.url/fails\": RoundTrip failed"))
		Expect(actual.LineNumber).Should(Equal(115))
		Expect(actual.Message).Should(Equal("API request failed; no response received"))
		Expect(actual.Package).Should(Equal("http"))
		Expect(actual.StatusCode).Should(BeZero())
		Expect(actual.Error()).Should(HaveSuffix("[test] http.WebClient.DoRequest (/goutils/http/client.go 115): " +
			"API request failed; no response received, Inner:\n\tGet \"test.url/fails\": RoundTrip failed."))
	})

	// Test that GetData will fail to get the data if the body fails to read
	It("GetData - GetBody fails - Error", func() {

		// First, create the test client
		httpClient := testutils.NewTestClient(false, func(req *http.Request) *http.Response {

			// Verify the request objects
			Expect(req.URL.String()).Should(Equal("test.url/fails"))
			Expect(req.Method).Should(Equal(http.MethodGet))

			// Return the response
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(testutils.ErrorReader(0)),
				Request:    req,
				Header:     make(http.Header),
			}
		})

		// Next, create the web client from the test client
		client := generateClient(httpClient)

		// Now, create the HTTP request
		request, _ := http.NewRequest(http.MethodGet, "test.url/fails", http.NoBody)
		request.Header.Add("Authorization", "Bearer FAKE_KEY")

		// Finally, attempt to send the request; this should fail
		var value test
		err := client.GetData(request, &value)
		actual := err.(*Error)

		// Verify the failure
		Expect(actual).Should(HaveOccurred())
		Expect(actual.Class).Should(Equal("WebClient"))
		Expect(actual.Environment).Should(Equal("test"))
		Expect(actual.File).Should(Equal("/goutils/http/client.go"))
		Expect(actual.Function).Should(Equal("GetBody"))
		Expect(actual.GeneratedAt).ShouldNot(BeNil())
		Expect(actual.Inner).Should(HaveOccurred())
		Expect(actual.Inner.Error()).Should(Equal("Read failed"))
		Expect(actual.LineNumber).Should(Equal(125))
		Expect(actual.Message).Should(Equal("Error reading response body"))
		Expect(actual.Package).Should(Equal("http"))
		Expect(actual.StatusCode).Should(BeZero())
		Expect(actual.Error()).Should(HaveSuffix("[test] http.WebClient.GetBody (/goutils/http/client.go 125): " +
			"Error reading response body, Inner:\n\tRead failed."))
	})

	// Test that GetData will fail to get the data if the payload fails to deserialize
	It("GetData - Deserialize fails - Error", func() {

		// First, create the test client
		httpClient := testutils.NewTestClient(false, testutils.VerifyAndGenerateResponse(http.MethodGet,
			"test.url/fails", http.StatusOK, "{\"Key\":\"herp\",\"Value\":\"derp\"}"))

		// Next, create the web client from the test client
		client := generateClient(httpClient)

		// Now, create the HTTP request
		request, _ := http.NewRequest(http.MethodGet, "test.url/fails", http.NoBody)
		request.Header.Add("Authorization", "Bearer FAKE_KEY")

		// Create a value that we can use to test that deserialization fails
		var value struct {
			Key   string
			Value int
		}

		// Finally, attempt to send the request; this should fail
		err := client.GetData(request, &value)
		actual := err.(*Error)

		// Verify the failure
		Expect(actual).Should(HaveOccurred())
		Expect(actual.Class).Should(Equal("WebClient"))
		Expect(actual.Environment).Should(Equal("test"))
		Expect(actual.File).Should(Equal("/goutils/http/client.go"))
		Expect(actual.Function).Should(Equal("Deserialize"))
		Expect(actual.GeneratedAt).ShouldNot(BeNil())
		Expect(actual.Inner).Should(HaveOccurred())
		Expect(actual.Inner.Error()).Should(Equal("json: cannot unmarshal string into Go struct field .Value of type int"))
		Expect(actual.LineNumber).Should(Equal(140))
		Expect(actual.Message).Should(Equal("Failed to unmarsahl JSON response body"))
		Expect(actual.Package).Should(Equal("http"))
		Expect(actual.StatusCode).Should(BeZero())
		Expect(actual.Error()).Should(HaveSuffix("[test] http.WebClient.Deserialize (/goutils/http/client.go 140): " +
			"Failed to unmarsahl JSON response body, Inner:\n\tjson: cannot unmarshal string into Go struct field " +
			".Value of type int."))
	})

	// Test that GetData will successfully ge the data if no errors occur
	It("GetData - No failures - Data populated", func() {

		// First, create the test client
		httpClient := testutils.NewTestClient(false, testutils.VerifyAndGenerateResponse(http.MethodGet,
			"test.url/fails", http.StatusOK, "{\"Key\":\"herp\",\"Value\":\"derp\"}"))

		// Next, create the web client from the test client
		client := generateClient(httpClient)

		// Now, create the HTTP request
		request, _ := http.NewRequest(http.MethodGet, "test.url/fails", http.NoBody)
		request.Header.Add("Authorization", "Bearer FAKE_KEY")

		// Finally, attempt to send the request; this should not fail
		var value test
		err := client.GetData(request, &value)

		// Verify the failure
		Expect(err).ShouldNot(HaveOccurred())
		Expect(value.Key).Should(Equal("herp"))
		Expect(value.Value).Should(Equal("derp"))
	})
})

// Helper function that generates a fake client that can be used for testing
func generateClient(client *http.Client) *WebClient {
	logger := utils.NewLogger("testd", "test")
	logger.Discard()
	pClient := WithClient(client, logger,
		WithRetryCodes([]int{http.StatusBadGateway, http.StatusRequestTimeout,
			http.StatusConflict, http.StatusTooManyRequests}),
		WithBackoffStart(1), WithBackoffEnd(5), WithBackoffMaxElapsed(10),
		WithErrorHandler(func(client *WebClient, data []byte) string { return "TEST ERROR" }))
	return pClient
}

// Test type that we'll use for serialization/deserialization tests
type test struct {
	Key   string
	Value string
}
