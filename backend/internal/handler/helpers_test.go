package handler

import (
	"encoding/json"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestRouter() *gin.Engine {
	return gin.New()
}

func assertErrorBody(t *testing.T, body []byte, wantCode string) {
	t.Helper()
	var e ErrorResponse
	if err := json.Unmarshal(body, &e); err != nil {
		t.Fatalf("invalid error JSON: %v (body: %s)", err, body)
	}
	if e.Error != wantCode {
		t.Errorf("error code = %q, want %q", e.Error, wantCode)
	}
	if e.Message == "" {
		t.Error("error message is empty")
	}
}
