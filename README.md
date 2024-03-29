[![Unit Tests Status](https://github.com/xefino/goutils/actions/workflows/test.yml/badge.svg)](https://github.com/xefino/goutils/actions)

# goutils
Common Go codebase for Xefino projects. This repo has been open-sourced for the purpose of rendering a service to the wider Go community and ensuring that potential errors and bugs in our own projects are surfaced more easily.

## Installation
Installing this repo is as simple as cloning it. However, there is some additional setup necessary in order to get things in working order.

### Golang Settings
This package was developed using Golang version 1.18. Lower versions will fail to compile. Moreover, this module was setup with the following golang setup:

```
GO111MODULE="on"
GOFLAGS="-mod=on"
```

We have this setup so that the code included in `vendor` will be downloaded with the repo, which makes deployuments simpler than they would be otherwise. When adding new packages to the repo, you should do `go get -u {repo}` and then call `go mod vendor` to ensure that the code is downloaded into the `vendor` directory.

## Repository Structure
Currently, the repo is structured as a number of directories based on functionality. Each directory is described below:

- /auth: Contains helper functions useful in authentication
- /awssvc: Contains standardized request functionality for AWS services, including DynamoDB, KMS, Lambda, S3, SQS. This library also includes an emulation of the AWS policy document that can be serialized to JSON.
- /collections: Contains functions and data structure types that are commonly used throughout Xefino projects.
- /concurrency: Convenience functions that make handling concurrent workloads easier
- /extensions: Extension functionality for various Go types
- /http: Standardized HTTP client for working with REST APIs
- /math: Arithmetic and logical helper functions
- /random: Functions to generate random data
- /reflection: Functions used to handle run-time reflection, such as getting all the field info from an object.
- /servicehelpers: Functions used to help set up service scheduling infrastructure
- /sql: Helper functions for SQL-related code, such as reading results from an SQL query into a list of objects. This library also contains some ORM helper functions.
- /strings: Utility functions for string manipulation
- /testutils: Utility functions commonly used in testing
- /time: Helper functions used when working with `time.Time` objects
- /utils: Contains utility types such as loggers and error handlers.
- /vendor: External modules installed by `go get`. **Do not make manual changes to this directory**

# Development
Please read and adhere to these development [guidelines](CONTRIBUTING.md) when contributing to this repo.

See our [wiki](https://github.com/xefino/goutils/wiki) here.
