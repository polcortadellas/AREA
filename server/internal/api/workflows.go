// Package api holds the HTTP handlers of the api role.
package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/polcortadellas/AREA/server/internal/node"
)

const maxBody = 1 << 20

func CreateWorkflow(reg *node.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxBody)
		var wf node.Workflow
		if err := json.NewDecoder(r.Body).Decode(&wf); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_body"})
			return
		}
		var invalid *node.InvalidWorkflowError
		if err := reg.ValidateWorkflow(wf); errors.As(err, &invalid) {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]any{
				"error":    "invalid_workflow",
				"problems": invalid.Problems,
			})
			return
		} else if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_error"})
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
