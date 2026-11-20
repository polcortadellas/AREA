// Package api holds the OpenAPI contract, embedded so the binary can serve it.
package api

import _ "embed"

//go:embed openapi.yaml
var Spec []byte
