package api

import (
	"encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/polcortadellas/AREA/server/internal/node"
)

// Bump when a change to /about.json would break existing clients.
const aboutSchemaVersion = 1

type about struct {
	Client aboutClient `json:"client"`
	Server aboutServer `json:"server"`
}

type aboutClient struct {
	Host string `json:"host"`
}

type aboutServer struct {
	SchemaVersion int            `json:"schema_version"`
	CurrentTime   int64          `json:"current_time"`
	Services      []aboutService `json:"services"`
}

type aboutService struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	OAuth2      bool        `json:"oauth2"`
	Actions     []aboutNode `json:"actions"`
	Reactions   []aboutNode `json:"reactions"`
}

type aboutNode struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Inputs      node.Schema `json:"inputs"`
	Outputs     node.Schema `json:"outputs"`
}

// About reads everything from the registry, so a new node appears here
// without editing this file.
func About(reg *node.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}

		services := make([]aboutService, 0)
		for _, s := range reg.Services() {
			services = append(services, aboutService{
				Name:        s.Name,
				Description: s.Description,
				OAuth2:      s.OAuth2,
				Actions:     describe(reg.Triggers(s.Name)),
				Reactions:   describe(reg.Reactions(s.Name)),
			})
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(about{
			Client: aboutClient{Host: host},
			Server: aboutServer{
				SchemaVersion: aboutSchemaVersion,
				CurrentTime:   time.Now().Unix(),
				Services:      services,
			},
		})
	}
}

func describe[N node.Node](nodes []N) []aboutNode {
	out := make([]aboutNode, 0, len(nodes))
	for _, n := range nodes {
		m := n.Metadata()
		out = append(out, aboutNode{
			Name:        m.Name,
			Type:        m.ID(),
			Description: m.Description,
			Inputs:      orEmpty(n.InputSchema()),
			Outputs:     orEmpty(n.OutputSchema()),
		})
	}
	return out
}

// Clients loop over inputs and outputs, so send [] instead of null.
func orEmpty(s node.Schema) node.Schema {
	if s == nil {
		return node.Schema{}
	}
	return s
}
