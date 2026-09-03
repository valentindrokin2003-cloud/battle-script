package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ReadyzHandler checks a real dependency (the database), unlike
// /healthz which is a pure liveness check with no external calls.
type ReadyzHandler struct {
	DB *sql.DB
}

func (h ReadyzHandler) Check(c *gin.Context) {
	if err := h.DB.PingContext(c.Request.Context()); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready", "reason": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}
