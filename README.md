# go-tautulli #

[![Test Status](https://github.com/amphitheaterr/go-tautulli/workflows/tests/badge.svg)](https://github.com/amphitheaterr/go-tautulli/actions?query=workflow%3Atests)
[![Test Coverage](https://codecov.io/gh/amphitheaterr/go-tautulli/branch/main/graph/badge.svg?token=Fi1pmOMEvK)](https://codecov.io/gh/amphitheaterr/go-tautulli)
[![Go Reference](https://pkg.go.dev/badge/github.com/amphitheaterr/go-tautulli.svg)](https://pkg.go.dev/github.com/amphitheaterr/go-tautulli)

go-tautulli is a Go client library for accessing the [Tautulli API v2][].

Currently, **go-tautulli requires Go version 1.17 or greater**.  go-github tracks
[Go's version support policy][support-policy].  We do our best not to break
older versions of Go if we don't have to, but due to tooling constraints, we
don't always test older versions.

[support-policy]: https://golang.org/doc/devel/release.html#policy

## Installation ##

go-tautulli is compatible with modern Go releases in module mode, with Go installed:

```bash
go get github.com/amphitheaterr/go-tautulli/v1
```
will resolve and add the package to the current development module, along with its dependencies.

Alternatively the same can be achieved if you use import in a package:

```go
import "github.com/amphitheaterr/go-tautulli/v1/github"
```

and run `go get` without parameters.

Finally, to use the top-of-trunk version of this repo, use the following command:

```bash
go get github.com/amphitheaterr/go-tautulli/v1@main
```

[Tautulli API v2]: https://github.com/Tautulli/Tautulli/wiki/Tautulli-API-Reference
