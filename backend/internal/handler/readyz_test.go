package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/valentindrokin2003-cloud/battle-script/backend/internal/testutil"
)

func TestReadyzHandler_DatabaseUp(t *testing.T) {
	db := testutil.OpenTestDB(t)
	defer func() { _ = db.Close() }()

	r := newTestRouter()
	r.GET("/readyz", ReadyzHandler{DB: db}.Check)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
}

func TestReadyzHandler_DatabaseDown(t *testing.T) {
	db := testutil.OpenTestDB(t)
	_ = db.Close() // deliberately closed before use

	r := newTestRouter()
	r.GET("/readyz", ReadyzHandler{DB: db}.Check)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503; body: %s", rec.Code, rec.Body.String())
	}
}
