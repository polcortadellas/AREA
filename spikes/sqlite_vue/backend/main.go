package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

type note struct {
	ID   int64  `json:"id"`
	Text string `json:"text"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func newHandler(db *sql.DB) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /notes", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, text FROM notes ORDER BY id")
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		defer rows.Close()
		notes := []note{}
		for rows.Next() {
			var n note
			if err := rows.Scan(&n.ID, &n.Text); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			notes = append(notes, n)
		}
		writeJSON(w, http.StatusOK, notes)
	})
	mux.HandleFunc("POST /notes", func(w http.ResponseWriter, r *http.Request) {
		var in note
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil || strings.TrimSpace(in.Text) == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "text required"})
			return
		}
		in.Text = strings.TrimSpace(in.Text)
		res, err := db.Exec("INSERT INTO notes (text) VALUES (?)", in.Text)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		in.ID, _ = res.LastInsertId()
		writeJSON(w, http.StatusCreated, in)
	})
	return mux
}

func openDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	// one connection: ":memory:" is per connection, and SQLite has a single writer anyway
	db.SetMaxOpenConns(1)
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS notes (id INTEGER PRIMARY KEY, text TEXT NOT NULL)")
	return db, err
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	path := os.Getenv("DB_PATH")
	if path == "" {
		path = "spike.db"
	}
	db, err := openDB(path)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(":"+port, newHandler(db)))
}
