// Package sisyphus holds what the sisyphus binaries share at build time.
package sisyphus

import (
	_ "embed"
	"strings"
)

//go:embed VERSION
var version string

// Version is the version the binary was built from: the VERSION file at the root of this module,
// embedded at compile time. It does not depend on the repo the binary later runs in.
func Version() string {
	return strings.TrimSpace(version)
}
