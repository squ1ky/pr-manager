package v1

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
)

type ErrorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
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

func newErrorResponseWithDetails(code, message string, details map[string]string) ErrorResponse {
	return ErrorResponse{
		Error: ErrorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

func bindAndValidate(c *gin.Context, dst any) bool {
	if err := c.ShouldBind(dst); err != nil {
		var verr validator.ValidationErrors
		if errors.As(err, &verr) {
			details := make(map[string]string, len(verr))
			for _, fe := range verr {
				field := fe.Field()
				details[field] = validationMessage(fe)
			}

			validationError(c, details)
			return false
		}

		respondError(c, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body")
		return false
	}

	return true
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

func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "field is required"
	case "max":
		return "too long (max " + fe.Param() + ")"
	case "min":
		return "too short (min " + fe.Param() + ")"
	case "notblank":
		return "must not be blank"
	default:
		return "invalid value"
	}
}

func validationError(c *gin.Context, details map[string]string) {
	c.JSON(http.StatusUnprocessableEntity, newErrorResponseWithDetails("VALIDATION_ERROR", "invalid request payload", details))
}

func internalError(c *gin.Context) {
	respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
}
