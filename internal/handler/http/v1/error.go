package v1

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

func newErrorResponse(code, message string) ErrorResponse {
	return ErrorResponse{
		Error: ErrorBody{
			Code:    code,
			Message: message,
		},
	}
}

func respondError(c *gin.Context, status int, code, message string) {
	c.JSON(status, newErrorResponse(code, message))
}

func badRequest(c *gin.Context, code, message string) {
	respondError(c, http.StatusBadRequest, code, message)
}

func notFound(c *gin.Context, message string) {
	respondError(c, http.StatusNotFound, "NOT_FOUND", message)
}

func internalError(c *gin.Context) {
	respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
}
