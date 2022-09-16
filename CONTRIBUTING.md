# How to contribute
Welcome to the goutils repo. Here are a few guidelines that we need contributors to follow to make our development experience a little bit less messy.

Generally, we care about:

- Readability, the most important.
- Maintainability. We avoid the code that surprises. Clear over clever.

# Conventions
- Make sure to check these guidelines:
  - [https://golang.org/doc/code.html](https://golang.org/doc/code.html)
  - [https://golang.org/doc/effective_go.html](https://golang.org/doc/effective_go.html)
- Use [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports?utm_source=godoc) to format your code. It's a good idea to run the tool on file save. Try to set this option (if available) in your preferred editor. We prefer goimports vs gofmt.
- Use lowercase for your package names, i.e. in pkg/, or services/, etc.
- Use lowercase for your filenames as well.
- Do not use underscore (\_) in your filenames except `_test.go`. For example, it should be `batchoperation.go`, not `batch_operation.go`.

# Logging
We have a logging library in `/utils/logging.go` and this should be used. If additional functionality is necessary, it can be extended. We do this to ensure that all our logs have a consistent format and feel. This standard should be adhered to unless an exception, valid reason can be given. This library can also be used to generate errors that have a standard format:

```
logger := utils.NewLogger("myserviced", os.Getenv("ENVIRONMENT"))
logger.Log("This is a backend log message. It can be used similarly to fmt.Sprintf. " + 
  "Here's a string arg: %s, an integer arg: %d, a float arg: %f", "ice cream", 42, 13.37)
err := logger.Error(inner_err, "Oh no, something bad happend: integer arg=%d", 42)
```

# Testing
Make sure all tests succeed in the root for your changes. You can run all tests in your root with the following command:

```
go test -v ./... -count=1 -cover -race -vet=off -coverprofile cover.out
```

We expect that test coverage should never drop below 85% at the package level. If this level of coverage cannot be maintained then that probably indicates that the module was not written in an easily-testable manner. Remember, untested code is broken code.
