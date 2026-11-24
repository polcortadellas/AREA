// Package github is the GitHub connector. Users link their account with OAuth2
// (internal/oauth); the trigger and the reaction come with M-15.
package github

import "github.com/polcortadellas/AREA/server/internal/node"

var Service = node.Service{
	Name:        "github",
	Description: "GitHub repositories and issues.",
	OAuth2:      true,
}
