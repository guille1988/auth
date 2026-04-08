package stress

import (
	"auth/tests/integration"
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
)

func TestStressModule(test *testing.T) {
	integration.TestCase(test, "it should return accepted when sending a stress email", func(test *testing.T) {
		payload := map[string]string{
			"email": gofakeit.Email(),
			"name":  gofakeit.Name(),
		}
		body, _ := json.Marshal(payload)

		request, _ := http.NewRequest("POST", "/api/stress", bytes.NewBuffer(body))
		request.Header.Set("Content-Type", "application/json")

		response := integration.ExecuteRequest(request)

		assert.Equal(test, http.StatusAccepted, response.Code)
	})

	integration.TestCase(test, "it should return validation error when email is missing", func(test *testing.T) {
		payload := map[string]string{
			"name": gofakeit.Name(),
		}
		body, _ := json.Marshal(payload)

		request, _ := http.NewRequest("POST", "/api/stress", bytes.NewBuffer(body))
		request.Header.Set("Content-Type", "application/json")

		response := integration.ExecuteRequest(request)

		assert.Equal(test, http.StatusUnprocessableEntity, response.Code)
	})

	integration.TestCase(test, "it should return validation error when email is invalid", func(test *testing.T) {
		payload := map[string]string{
			"email": "not-a-valid-email",
			"name":  gofakeit.Name(),
		}
		body, _ := json.Marshal(payload)

		request, _ := http.NewRequest("POST", "/api/stress", bytes.NewBuffer(body))
		request.Header.Set("Content-Type", "application/json")

		response := integration.ExecuteRequest(request)

		assert.Equal(test, http.StatusUnprocessableEntity, response.Code)
	})
}
