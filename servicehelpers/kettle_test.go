package servicehelpers

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/gomodule/redigo/redis"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/xefino/goutils/collections"
	"github.com/xefino/goutils/utils"
)

var _ = Describe("Kettle Context Tests", func() {

	// Tests that the NewKettleContext function fails if the REDIS_HOST environment variable is not set
	It("NewKettleContext - REDIS_HOST not set - Error", func() {

		// First, unset the REDIS_HOST variable so that we can test around the create-context function
		// fails. Ensure that it is reset after the test finishes running
		redisHost, ok := os.LookupEnv("REDIS_HOST")
		if ok {
			os.Unsetenv("REDIS_HOST")
			defer os.Setenv("REDIS_HOST", redisHost)
		}

		// Next, create our test logger and ensure that it doesn't send anything to the standard output
		logger := utils.NewLogger("testd", "test")
		logger.Discard()

		// Now, attempt to create a Kettle context from a service name and environment; this should not fail
		serviceName := fmt.Sprintf("testd_%d", rand.Int())
		kCtx, err := NewKettleContext(serviceName, "test", 30, logger)
		Expect(err).Should(HaveOccurred())

		// Finally, verify the error
		actual := err.(*utils.GError)
		Expect(kCtx).Should(BeNil())
		Expect(actual.Class).Should(BeEmpty())
		Expect(actual.Environment).Should(Equal("test"))
		Expect(actual.File).Should(Equal("/goutils/servicehelpers/kettle.go"))
		Expect(actual.Function).Should(Equal("NewKettleContext"))
		Expect(actual.GeneratedAt).ShouldNot(BeNil())
		Expect(actual.Inner).Should(HaveOccurred())
		Expect(actual.Inner.Error()).Should(Equal("NewRedisPool failed: REDIS_HOST not set (host:port)"))
		Expect(actual.LineNumber).Should(Equal(58))
		Expect(actual.Message).Should(Equal("Failed to create new Kettle instance"))
		Expect(actual.Package).Should(Equal("servicehelpers"))
		Expect(actual.Error()).Should(HaveSuffix("[test] servicehelpers.NewKettleContext (/goutils/servicehelpers/kettle.go 58): " +
			"Failed to create new Kettle instance, Inner:\n\tNewRedisPool failed: REDIS_HOST not set (host:port)."))
	})

	// Tests that the NewKettleContext function works as expected
	It("NewKettleContext - Works", func() {

		// First, ensure that the REDIS_HOST environment variable has been set if it isn't set already.
		// Ensure that the environment is returned to the same condition when the test finishes
		_, ok := os.LookupEnv("REDIS_HOST")
		if !ok {
			os.Setenv("REDIS_HOST", "localhost:6379")
			defer os.Unsetenv("REDIS_HOST")
		}

		// Next, create our test logger and ensure that it doesn't send anything to the standard output
		logger := utils.NewLogger("testd", "test")
		logger.Discard()

		// Now, attempt to create a Kettle context from a service name and environment; this should not fail
		serviceName := fmt.Sprintf("testd_%d", rand.Int())
		kCtx, err := NewKettleContext(serviceName, "test", 30, logger)
		Expect(err).ShouldNot(HaveOccurred())

		// Finally, verify that the context was created correctly
		Expect(kCtx.Name).Should(Equal(serviceName))
		Expect(kCtx.Environment).Should(Equal("test"))
		Expect(kCtx.Instance).ShouldNot(BeNil())
		Expect(kCtx.Logger).ShouldNot(BeNil())
	})

	// Tests the conditions under which the Run function could fail and that none of these failure
	// modes actual cause Run to fail
	DescribeTable("Run - Failures",
		func(getFails bool, setFails bool, operationFails bool) {

			// First, ensure that the REDIS_HOST environment variable has been set if it isn't set already.
			// Ensure that the environment is returned to the same condition when the test finishes
			_, ok := os.LookupEnv("REDIS_HOST")
			if !ok {
				os.Setenv("REDIS_HOST", "localhost:6379")
				defer os.Unsetenv("REDIS_HOST")
			}

			// Next, create our test logger and ensure that it doesn't send anything to the standard output
			logger := utils.NewLogger("testd", "test")
			logger.Discard()

			// Now, attempt to create a Kettle context from a service name and environment; this should not fail
			serviceName := fmt.Sprintf("testd_%d", rand.Int())
			kCtx, err := NewKettleContext(serviceName, "test", 30, logger)
			Expect(err).ShouldNot(HaveOccurred())

			// Ensure that Redis is cleared of test data after the test concludes
			key := fmt.Sprintf("test-%t-%t-%t", getFails, setFails, operationFails)
			fullKey := fmt.Sprintf("%s_%s", serviceName, key)
			defer func() {
				_, rErr := redis.Int(kCtx.Instance.Pool().Get().Do("DEL", fullKey))
				Expect(rErr).ShouldNot(HaveOccurred())
			}()

			// If we want to test around GET failing then set an invalid value to the key here
			if getFails {
				kCtx.Instance.Pool().Get().Do("SET", fullKey, "derp")
			}

			// If we want to test around SETEX failing then set an invalid expiry here
			freq := time.Second
			if setFails {
				freq = time.Nanosecond
			}

			// Finally, attempt to start the Kettle instance with a master and slave function
			err = kCtx.Run(context.Background(),
				func(ctx context.Context) error {
					time.Sleep(200 * time.Millisecond)
					return nil
				},
				NewKeyer(key, freq,
					func(ctx context.Context) error {
						if operationFails {
							return kCtx.Logger.Error(nil, "Operation failed")
						}

						return nil
					}))

			// Verify the details of the error
			Expect(err).ShouldNot(HaveOccurred())
		},
		Entry("GET fails - Error", true, false, false),
		Entry("SETEX fails - Error", false, true, false),
		Entry("Operation fails - Error", false, false, true))

	// Test that the Run function works as expected
	It("Run - Works", func() {

		// First, ensure that the REDIS_HOST environment variable has been set if it isn't set already.
		// Ensure that the environment is returned to the same condition when the test finishes
		_, ok := os.LookupEnv("REDIS_HOST")
		if !ok {
			os.Setenv("REDIS_HOST", "localhost:6379")
			defer os.Unsetenv("REDIS_HOST")
		}

		// Next, create our test logger and ensure that it doesn't send anything to the standard output
		logger := utils.NewLogger("testd", "test")
		logger.Discard()

		// Now, attempt to create a Kettle context from a service name and environment; this should not fail
		serviceName := fmt.Sprintf("testd_%d", rand.Int())
		kCtx, err := NewKettleContext(serviceName, "test", 2, logger)
		Expect(err).ShouldNot(HaveOccurred())

		// Ensure that Redis is cleared of test data after the test concludes
		defer func() {
			_, rErr := redis.Int(kCtx.Instance.Pool().Get().Do("DEL", fmt.Sprintf("%s_test", serviceName)))
			Expect(rErr).ShouldNot(HaveOccurred())
		}()

		// Finally, attempt to start the Kettle instance with a master and slave function
		results := collections.NewConcurrentList[string]()
		err = kCtx.Run(context.Background(),
			func(ctx context.Context) error {
				time.Sleep(3 * time.Second)
				results.Append("SLAVE")
				return nil
			},
			NewKeyer("test", time.Minute, func(ctx context.Context) error {
				results.Append("MASTER")
				return nil
			}))

		// Verify the results
		Expect(err).ShouldNot(HaveOccurred())
		Expect(results.Length()).Should(Equal(uint(2)))
		Expect(results.Clear()).Should(ContainElements("MASTER", "SLAVE"))
	})
})
