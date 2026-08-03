// Command fugovet is a go vet-compatible static analysis tool for Fugo apps.
//
// Usage:
//
//	go build -o fugovet ./cmd/fugovet
//	go vet -vettool=$(which fugovet) ./...
//
// See github.com/sazardev/fugo/fugovet for the individual checks.
package main

import (
	"golang.org/x/tools/go/analysis/multichecker"

	"github.com/sazardev/fugo/fugovet"
)

func main() {
	multichecker.Main(fugovet.Analyzers...)
}
