package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/polcortadellas/AREA/server/internal/auth"
)

type RunQueue interface {
	Enqueue(ctx context.Context, runID uuid.UUID) error
}

var errWorkflowNotFound = errors.New("workflow not found")

// Publishes after the commit: a failed publish leaves the run PENDING for the recovery sweep.
func TriggerWorkflow(pool *pgxpool.Pool, queue RunQueue) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := auth.UserID(r.Context())
		workflowID, err := uuid.Parse(r.PathValue("workflowId"))
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
			return
		}

		var runID uuid.UUID
		err = pgx.BeginFunc(r.Context(), pool, func(tx pgx.Tx) error {
			var exists bool
			if err := tx.QueryRow(r.Context(),
				`SELECT EXISTS (SELECT 1 FROM workflows WHERE id = $1 AND user_id = $2)`, workflowID, userID,
			).Scan(&exists); err != nil {
				return err
			}
			if !exists {
				return errWorkflowNotFound
			}
			// No external event: each manual trigger gets its own ID.
			return tx.QueryRow(r.Context(),
				`INSERT INTO runs (workflow_id, external_event_id) VALUES ($1, $2) RETURNING id`,
				workflowID, uuid.NewString(),
			).Scan(&runID)
		})
		switch {
		case errors.Is(err, errWorkflowNotFound):
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
			return
		case err != nil:
			slog.Error("trigger: store run", "err", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_error"})
			return
		}

		if err := queue.Enqueue(r.Context(), runID); err != nil {
			slog.Warn("trigger: enqueue failed, the recovery sweep will retry", "run", runID, "err", err)
		}
		writeJSON(w, http.StatusAccepted, map[string]string{"id": runID.String()})
	}
}
