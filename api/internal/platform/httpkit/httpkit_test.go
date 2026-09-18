package httpkit

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestIDAndErrorEnvelopeAreConsistent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestIDMiddleware())
	var contextRequestID string
	router.GET("/failure", func(c *gin.Context) {
		contextRequestID = RequestIDFromContext(c.Request.Context())
		WriteError(c, http.StatusUnprocessableEntity, Error{
			Code:    "VALIDATION_FAILED",
			Message: "Request fields are invalid",
			Details: []Detail{{Field: "email", Reason: "invalid_format"}},
		})
	})

	request := httptest.NewRequest(http.MethodGet, "/failure", nil)
	request.Header.Set("X-Request-ID", "caller-controlled\r\ninjected")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnprocessableEntity)
	}
	requestID := response.Header().Get("X-Request-ID")
	if requestID == "" || requestID == "caller-controlled\r\ninjected" {
		t.Fatalf("response request id was not server generated: %q", requestID)
	}
	if contextRequestID != requestID {
		t.Fatalf("context request id = %q, want %q", contextRequestID, requestID)
	}
	var body errorEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error.RequestID != requestID || body.Error.Code != "VALIDATION_FAILED" {
		t.Fatalf("error envelope request id/code mismatch: %+v", body.Error)
	}
	if len(body.Error.Details) != 1 || body.Error.Details[0].Field != "email" || body.Error.Details[0].Reason != "invalid_format" {
		t.Fatalf("error details did not preserve field and reason independently: %+v", body.Error.Details)
	}
}
