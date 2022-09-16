package utils

import (
	"fmt"
	"path"
	"runtime"
	"strings"
	"time"

	ustr "github.com/xefino/goutils/strings"
)

// ErrorProvider is a type that can be used when generating
// errors to ensure that the error is generated in the proper
// context
type ErrorProvider struct {
	SkipFrames  int
	PackageBase string
}

// DefaultErrorProvider contains the default settings to use when
// generating errors
var DefaultErrorProvider = ErrorProvider{SkipFrames: 2, PackageBase: "goutils"}

// GError contains information necessary to represent an error. We use the
// value GError instead of Error so, if it is embedded into another type, it
// will not invalidate the error interface
type GError struct {
	Environment string
	Package     string
	Class       string
	Function    string
	File        string
	LineNumber  int
	GeneratedAt time.Time
	Message     string
	Inner       error
}

// NewError creates a new error in the default context. See documentation
// of GenerateError for more information.
func NewError(env string, inner error, message string, args ...interface{}) *GError {
	return DefaultErrorProvider.GenerateError(env, inner, message, args...)
}

// GenerateBackendError creates a new error from an environment variable, an inner
// error, a message and a list of arguments to inject into the message. Note that
// this function assumes that the caller generated this error. If that is not the
// case then a custom error provider should be used with the proper number of skip
// frames. See the runtime documentation for more information.
func (provider ErrorProvider) GenerateError(env string, inner error,
	message string, args ...interface{}) *GError {

	// First, get the file and line number of the caller (we'll use this
	// to get information about where the error occurred) and then get
	// information about the function that generated the error
	ptr, file, line, _ := runtime.Caller(provider.SkipFrames)
	funcObj := runtime.FuncForPC(ptr)

	// Split the file by the string "vendor" so we can handle cases where a vendor
	// package produces this error (this will override the normal package rules)
	splitByVendor := strings.SplitAfter(file, "vendor")
	if len(splitByVendor) > 1 {
		file = splitByVendor[1]
	}

	// In order to avoid printing the path all the way to the root, let's
	// split the path at a set package base and then join after it so we can
	// strip off anything that's not really necessary to describe the file
	splitFile := strings.SplitAfter(file, provider.PackageBase)
	if len(splitFile) > 1 {
		file = fmt.Sprintf("/%s%s", provider.PackageBase, splitFile[len(splitFile)-1])
	}

	// Next, get the name of the function that was associated with the caller
	// We expect it to be either package.class.func or package.func so we'll split
	// the name by period as well. Also, if we have a generic type name then we
	// expect it will contain [...] so remove that string as well
	var class, name string
	fullName := strings.Replace(funcObj.Name(), "[...]", "", 1)
	splitName := strings.Split(fullName, ".")

	// Now, get the package name, class name (if present) and function name
	_, packageName := path.Split(splitName[1])
	if len(splitName) >= 4 {
		class = splitName[2]
		name = splitName[3]
	} else {
		name = splitName[2]
	}

	// If the class represents a pointer then we'll want to extract the
	// name from the class object so do that here
	class = strings.TrimRight(strings.TrimLeft(class, "(*"), ")")

	// Finally, inject all this information into the error itself
	// and return it
	return &GError{
		Environment: env,
		Package:     packageName,
		Class:       class,
		Function:    name,
		File:        file,
		LineNumber:  line,
		GeneratedAt: time.Now().UTC(),
		Message:     fmt.Sprintf(message, args...),
		Inner:       inner,
	}
}

// Error creates an error string from the backend error
func (err *GError) Error() string {

	// First, recombine the package, class and function name
	// into a fully-qualified function name
	var fullyQualifiedName string
	if ustr.IsEmpty(err.Class) {
		fullyQualifiedName = fmt.Sprintf("%s.%s", err.Package, err.Function)
	} else {
		fullyQualifiedName = fmt.Sprintf("%s.%s.%s", err.Package, err.Class, err.Function)
	}

	// Next, create the base message from the error data
	baseMsg := fmt.Sprintf("%s [%s] %s (%s %d): ", err.GeneratedAt, err.Environment, fullyQualifiedName,
		err.File, err.LineNumber) + err.Message

	// Finally, if the inner error isn't nil then add it to the error
	// message; otherwise, just add a period
	if err.Inner != nil {
		baseMsg += fmt.Sprintf(", Inner: %v.", err.Inner)
	} else {
		baseMsg += "."
	}

	return baseMsg
}

// AggregateError is a wrapper for multiple errors
type AggregateError []error

// FromErrors creates a new AggregateError from
func FromErrors(errs ...error) error {
	if len(errs) > 0 {
		err := AggregateError(errs)
		return err
	}

	return nil
}

// Error generates an error string from an aggregate error
func (err AggregateError) Error() string {

	// Iterate over all the inner errors and generate a message for each
	msgs := make([]string, len(err))
	for i, err := range err {
		msgs[i] = err.Error()
	}

	// Create a new message from the error messages and return it
	return fmt.Sprintf("Multiple errors occurred: \n\t%s", strings.Join(msgs, "\n\t"))
}

// As allows for conversion from an error to its actual type
func As[T error](inner error) T {
	return inner.(T)
}
