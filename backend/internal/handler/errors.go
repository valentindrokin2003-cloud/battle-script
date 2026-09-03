// Package handler is the Gin delivery layer: HTTP input validation,
// calls into internal/service, response serialization. No business
// logic lives here — see docs/superpowers/specs/2026-09-03-http-api-design.md.
package handler

import "github.com/gin-gonic/gin"

// ErrorResponse is the API's single error body shape, per the HTTP API
// spec's "Обработка ошибок".
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, ErrorResponse{Error: code, Message: message})
}
