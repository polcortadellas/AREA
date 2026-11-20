// Package api holds the HTTP handlers of the api role.
package api

import (
	"encoding/json"
	"net/http"

	"github.com/polcortadellas/AREA/server/internal/node"
)

const maxBody = 1 << 20

// shortcut: validates only, nothing is saved until M-08 gives us the user_id.
func CreateWorkflow(reg *node.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBody)
		var wf node.Workflow
		if err := json.NewDecoder(r.Body).Decode(&wf); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if err := reg.ValidateWorkflow(wf); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}
