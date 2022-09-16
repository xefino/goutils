package utils

import (
	"bytes"
	"fmt"
	"log"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logger Tests", func() {

	// Tests that creating a logger with a different call frame
	// works as expected
	It("ChangeFrame - Works", func() {

		// Create our logger from a test service with a test environment
		// and then generate a second logger with a different number of skip-frames
		logger1 := NewLogger("testd", "test",
			WithInfoLog(*log.New(os.Stdout, "info", log.LstdFlags)),
			WithErrorLog(*log.New(os.Stderr, "error", log.LstdFlags)),
			WithErrorProvider(ErrorProvider{SkipFrames: 2, PackageBase: "test"}),
			WithPrefix("derp"))
		logger2 := logger1.ChangeFrame(5)

		// Verify that the loggers are separate objects
		Expect(logger1.Prefix).Should(Equal("derp"))
		Expect(logger1.Environment).Should(Equal("test"))
		Expect(logger1.infoLog.Writer()).Should(Equal(os.Stdout))
		Expect(logger1.infoLog.Prefix()).Should(Equal("info"))
		Expect(logger1.infoLog.Flags()).Should(Equal(log.LstdFlags))
		Expect(logger1.errLog.Writer()).Should(Equal(os.Stderr))
		Expect(logger1.errLog.Prefix()).Should(Equal("error"))
		Expect(logger1.errLog.Flags()).Should(Equal(log.LstdFlags))
		Expect(logger1.errProvider.SkipFrames).Should(Equal(2))
		Expect(logger1.errProvider.PackageBase).Should(Equal("test"))
		Expect(logger2.Prefix).Should(Equal("derp"))
		Expect(logger2.Environment).Should(Equal("test"))
		Expect(logger2.infoLog.Writer()).Should(Equal(os.Stdout))
		Expect(logger2.infoLog.Prefix()).Should(Equal("info"))
		Expect(logger2.infoLog.Flags()).Should(Equal(log.LstdFlags))
		Expect(logger2.errLog.Writer()).Should(Equal(os.Stderr))
		Expect(logger2.errLog.Prefix()).Should(Equal("error"))
		Expect(logger2.errLog.Flags()).Should(Equal(log.LstdFlags))
		Expect(logger2.errProvider.SkipFrames).Should(Equal(5))
		Expect(logger2.errProvider.PackageBase).Should(Equal("test"))
	})

	// Tests that logging a message works as expected
	It("Log - Works", func() {

		// First, create our logger from a test service with a test environment
		logger := NewLogger("testd", "test")

		// Next, create a buffer and set the output to it so we can extract messages
		// from our logger
		buf := new(bytes.Buffer)
		logger.infoLog.SetOutput(buf)

		// Now, log a test message to the buffer
		logger.Log("Test message. String parameter: %s, Integer parameter: %d", "derp", 42)

		// Finally, extract the data from the buffer and verify the value of the message
		data := string(buf.Bytes())
		Expect(data).Should(HaveSuffix("[Info][test][testd] Test message. " +
			"String parameter: derp, Integer parameter: 42\n"))
	})

	// Tests that, even if one of the message paramters contains a control character, that
	// character will be printed as-is in the log
	It("Log - Message contains control character - Works", func() {

		// First, create our logger from a test service with a test environment
		logger := NewLogger("testd", "test")

		// Next, create a buffer and set the output to it so we can extract messages
		// from our logger
		buf := new(bytes.Buffer)
		logger.infoLog.SetOutput(buf)

		// Now, log a message where one of the parameters contains a control character
		logger.Log("Test message. String parameter: %q", "test %t.derp")

		// Finally, extract the data from the buffer and verify the value of the message
		data := string(buf.Bytes())
		Expect(data).Should(HaveSuffix("[Info][test][testd] Test message. " +
			"String parameter: \"test %t.derp\"\n"))
	})

	// Tests that logging an error works as expected
	It("Error - Works", func() {

		// First, create our logger from a test service with a test environment
		logger := NewLogger("testd", "test")

		// Next, create a buffer and set the output to it so we can extract messages
		// from our logger
		buf := new(bytes.Buffer)
		logger.errLog.SetOutput(buf)

		// Now, log an error message to the buffer
		err := logger.Error(fmt.Errorf("Test error"), "Test message. String parameter: %s, "+
			"Integer parameter: %d", "derp", 42)

		// Finally, extract the data from the buffer and verify the value of the message
		data := string(buf.Bytes())
		Expect(data).Should(HaveSuffix("[test] utils.glob. (/goutils/utils/logger_test.go 104): " +
			"Test message. String parameter: derp, Integer parameter: 42, Inner: Test error.\n"))

		// Verify the data in the error
		Expect(err.Class).Should(Equal("glob"))
		Expect(err.Environment).Should(Equal("test"))
		Expect(err.File).Should(Equal("/goutils/utils/logger_test.go"))
		Expect(err.Function).Should(BeEmpty())
		Expect(err.GeneratedAt).ShouldNot(BeNil())
		Expect(err.Inner).Should(HaveOccurred())
		Expect(err.Inner.Error()).Should(Equal("Test error"))
		Expect(err.LineNumber).Should(Equal(104))
		Expect(err.Message).Should(Equal("Test message. String parameter: derp, Integer parameter: 42"))
		Expect(err.Package).Should(Equal("utils"))
		Expect(err.Error()).Should(HaveSuffix("[test] utils.glob. (/goutils/utils/logger_test.go 104): " +
			"Test message. String parameter: derp, Integer parameter: 42, Inner: Test error."))
	})

	// Tests that the Discard function produces no data
	It("Discard - Works", func() {

		// First, create our logger from a test service with a test environment
		logger := NewLogger("testd", "test")

		// Next, create a buffer and set the output to it so we can extract messages
		// from our logger. Then call the discard function to ensure that the logger
		// output is discarded
		buf := new(bytes.Buffer)
		logger.infoLog.SetOutput(buf)
		logger.errLog.SetOutput(buf)
		logger.Discard()

		// Now, log a test message to the buffer and then generate an error
		logger.Log("Test message. String parameter: %s, Integer parameter: %d", "derp", 42)
		err := logger.Error(fmt.Errorf("Test error"), "Test message. String parameter: %s, "+
			"Integer parameter: %d", "derp", 42)

		// Verify the data in the error and that the buffer is empty
		Expect(buf.Bytes()).Should(BeEmpty())
		Expect(err.Class).Should(Equal("glob"))
		Expect(err.Environment).Should(Equal("test"))
		Expect(err.File).Should(Equal("/goutils/utils/logger_test.go"))
		Expect(err.Function).Should(BeEmpty())
		Expect(err.GeneratedAt).ShouldNot(BeNil())
		Expect(err.Inner).Should(HaveOccurred())
		Expect(err.Inner.Error()).Should(Equal("Test error"))
		Expect(err.LineNumber).Should(Equal(143))
		Expect(err.Message).Should(Equal("Test message. String parameter: derp, Integer parameter: 42"))
		Expect(err.Package).Should(Equal("utils"))
		Expect(err.Error()).Should(HaveSuffix("[test] utils.glob. (/goutils/utils/logger_test.go 143): " +
			"Test message. String parameter: derp, Integer parameter: 42, Inner: Test error."))
	})
})
