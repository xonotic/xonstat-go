//go:build tools

// Package tools pins the versions of command-line tools used by this project
// (e.g. swag) so that `go run` from within the module resolves the exact
// version recorded in go.mod / vendor/ the same way it does for libraries.
package main

import (
	_ "github.com/swaggo/swag/cmd/swag" // make swagger
)