# go-tautulli #

[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/cinehouse/go-tautulli)](https://github.com/cinehouse/go-tautulli/releases)
[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)](https://pkg.go.dev/github.com/cinehouse/go-tautulli/v1/tautulli)
[![GitHub Workflow Status (branch)](https://img.shields.io/github/workflow/status/cinehouse/go-tautulli/tests/main?label=tests)](https://github.com/cinehouse/go-tautulli/actions?query=workflow%3Atests)
[![Go Report Card](https://goreportcard.com/badge/github.com/cinehouse/go-tautulli)](https://goreportcard.com/report/github.com/cinehouse/go-tautulli)
[![Codecov branch](https://img.shields.io/codecov/c/github/cinehouse/go-tautulli/main?token=p78MbVUq1e)](https://codecov.io/gh/cinehouse/go-tautulli)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=cinehouse_go-tautulli&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=cinehouse_go-tautulli)
[![GitHub go.mod Go version (branch)](https://img.shields.io/github/go-mod/go-version/cinehouse/go-tautulli/main?label=Go)](https://golang.org/doc/install)
[![Go Reference](https://pkg.go.dev/badge/github.com/cinehouse/go-tautulli.svg)](https://pkg.go.dev/github.com/cinehouse/go-tautulli)
[![GitHub](https://img.shields.io/github/license/cinehouse/go-tautulli)](https://github.com/cinehouse/go-tautulli/blob/main/LICENSE)

go-tautulli is a Go client library for accessing the [Tautulli API v2][].

Currently, **go-tautulli requires Go version 1.17 or greater**.  go-github tracks
[Go's version support policy][support-policy].  We do our best not to break
older versions of Go if we don't have to, but due to tooling constraints, we
don't always test older versions.

[support-policy]: https://golang.org/doc/devel/release.html#policy

## Installation ##

go-tautulli is compatible with modern Go releases in module mode, with Go installed:

```bash
go get github.com/cinehouse/go-tautulli/v1
```
will resolve and add the package to the current development module, along with its dependencies.

Alternatively the same can be achieved if you use import in a package:

```go
import "github.com/cinehouse/go-tautulli/v1/tautulli"
```

and run `go get` without parameters.

Finally, to use the top-of-trunk version of this repo, use the following command:

```bash
go get github.com/cinehouse/go-tautulli/v1@main
```

[Tautulli API v2]: https://github.com/Tautulli/Tautulli/wiki/Tautulli-API-Reference
