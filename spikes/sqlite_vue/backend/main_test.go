package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func do(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, strings.NewReader(body)))
	return rec
}

func TestSpike(t *testing.T) {
	db, err := openDB(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	h := newHandler(db)

	if r := do(t, h, "GET", "/ping", ""); r.Code != 200 || strings.TrimSpace(r.Body.String()) != `{"status":"ok"}` {
		t.Fatalf("ping: %d %s", r.Code, r.Body)
	}
	if r := do(t, h, "POST", "/notes", `{"text":"hello"}`); r.Code != 201 {
		t.Fatalf("post: %d %s", r.Code, r.Body)
	}
	if r := do(t, h, "GET", "/notes", ""); strings.TrimSpace(r.Body.String()) != `[{"id":1,"text":"hello"}]` {
		t.Fatalf("get: %s", r.Body)
	}
	if r := do(t, h, "POST", "/notes", `{"text":" "}`); r.Code != 400 {
		t.Fatalf("empty note: %d", r.Code)
	}
}
