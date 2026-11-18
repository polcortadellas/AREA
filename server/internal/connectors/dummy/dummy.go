// Package dummy is a test service with no external calls, used to develop the
// engine before the real connectors exist.
package dummy

import "github.com/polcortadellas/AREA/server/internal/node"

var Service = node.Service{
	Name:        "dummy",
	Description: "Test service for development: makes no external calls.",
}
