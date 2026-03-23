package users

import (
	"auth/tests/integration"
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
)

func TestUserModule(test *testing.T) {
	integration.TestCase(test, "it should create a user", func(test *testing.T) {
		token := integration.GetToken()
		payload := map[string]string{
			"name":     gofakeit.Name(),
			"email":    gofakeit.Email(),
			"password": "password123",
		}
		body, _ := json.Marshal(payload)

		request, _ := http.NewRequest("POST", "/api/users/", bytes.NewBuffer(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+token)

		response := integration.ExecuteRequest(request)

		assert.Equal(test, http.StatusCreated, response.Code)
	})

	integration.TestCase(test, "it should list users.", func(test *testing.T) {
		token := integration.GetToken()
		request, _ := http.NewRequest("GET", "/api/users/", nil)
		request.Header.Set("Authorization", "Bearer "+token)
		response := integration.ExecuteRequest(request)

		assert.Equal(test, http.StatusOK, response.Code)
	})

	integration.TestCase(test, "it should show a user", func(test *testing.T) {
		token := integration.GetToken()
		userName := gofakeit.Name()
		userEmail := gofakeit.Email()
		payload := map[string]string{
			"name":     userName,
			"email":    userEmail,
			"password": "password123",
		}
		body, _ := json.Marshal(payload)

		createRequest, _ := http.NewRequest("POST", "/api/users/", bytes.NewBuffer(body))
		createRequest.Header.Set("Content-Type", "application/json")
		createRequest.Header.Set("Authorization", "Bearer "+token)
		integration.ExecuteRequest(createRequest)

		listRequest, _ := http.NewRequest("GET", "/api/users/", nil)
		listRequest.Header.Set("Authorization", "Bearer "+token)
		listResponse := integration.ExecuteRequest(listRequest)

		var users []map[string]any
		_ = json.Unmarshal(listResponse.Body.Bytes(), &users)

		var userUUID string
		for _, user := range users {
			if user["email"] == userEmail {
				userUUID = user["uuid"].(string)
				break
			}
		}

		showRequest, _ := http.NewRequest("GET", "/api/users/"+userUUID, nil)
		showRequest.Header.Set("Authorization", "Bearer "+token)
		showResponse := integration.ExecuteRequest(showRequest)

		var shownUser map[string]any
		_ = json.Unmarshal(showResponse.Body.Bytes(), &shownUser)

		assert.Equal(test, http.StatusOK, showResponse.Code)
		assert.Equal(test, userName, shownUser["name"])
		assert.Equal(test, userEmail, shownUser["email"])
	})

	integration.TestCase(test, "it should return 404 when user not found", func(test *testing.T) {
		token := integration.GetToken()
		request, _ := http.NewRequest("GET", "/api/users/"+gofakeit.UUID(), nil)
		request.Header.Set("Authorization", "Bearer "+token)
		response := integration.ExecuteRequest(request)

		assert.Equal(test, http.StatusNotFound, response.Code)
	})

	integration.TestCase(test, "it should update a user", func(test *testing.T) {
		token := integration.GetToken()
		userEmail := gofakeit.Email()
		payload := map[string]string{
			"name":     gofakeit.Name(),
			"email":    userEmail,
			"password": "password123",
		}
		body, _ := json.Marshal(payload)

		createRequest, _ := http.NewRequest("POST", "/api/users/", bytes.NewBuffer(body))
		createRequest.Header.Set("Content-Type", "application/json")
		createRequest.Header.Set("Authorization", "Bearer "+token)
		integration.ExecuteRequest(createRequest)

		listRequest, _ := http.NewRequest("GET", "/api/users/", nil)
		listRequest.Header.Set("Authorization", "Bearer "+token)
		listResponse := integration.ExecuteRequest(listRequest)

		var users []map[string]any
		_ = json.Unmarshal(listResponse.Body.Bytes(), &users)

		var userUUID string
		for _, user := range users {
			if user["email"] == userEmail {
				userUUID = user["uuid"].(string)
				break
			}
		}

		newName := gofakeit.Name()
		updatePayload := map[string]string{
			"name": newName,
		}
		updateBody, _ := json.Marshal(updatePayload)

		updateRequest, _ := http.NewRequest("PATCH", "/api/users/"+userUUID, bytes.NewBuffer(updateBody))
		updateRequest.Header.Set("Content-Type", "application/json")
		updateRequest.Header.Set("Authorization", "Bearer "+token)
		updateResponse := integration.ExecuteRequest(updateRequest)

		assert.Equal(test, http.StatusNoContent, updateResponse.Code)

		showRequest, _ := http.NewRequest("GET", "/api/users/"+userUUID, nil)
		showRequest.Header.Set("Authorization", "Bearer "+token)
		showResponse := integration.ExecuteRequest(showRequest)

		var shownUser map[string]any
		_ = json.Unmarshal(showResponse.Body.Bytes(), &shownUser)

		assert.Equal(test, newName, shownUser["name"])
	})

	integration.TestCase(test, "it should delete a user", func(test *testing.T) {
		token := integration.GetToken()
		userEmail := gofakeit.Email()
		payload := map[string]string{
			"name":     gofakeit.Name(),
			"email":    userEmail,
			"password": "password123",
		}
		body, _ := json.Marshal(payload)

		createRequest, _ := http.NewRequest("POST", "/api/users/", bytes.NewBuffer(body))
		createRequest.Header.Set("Content-Type", "application/json")
		createRequest.Header.Set("Authorization", "Bearer "+token)
		integration.ExecuteRequest(createRequest)

		listRequest, _ := http.NewRequest("GET", "/api/users/", nil)
		listRequest.Header.Set("Authorization", "Bearer "+token)
		listResponse := integration.ExecuteRequest(listRequest)

		var users []map[string]any
		_ = json.Unmarshal(listResponse.Body.Bytes(), &users)

		var userUUID string
		for _, user := range users {
			if user["email"] == userEmail {
				userUUID = user["uuid"].(string)
				break
			}
		}

		deleteRequest, _ := http.NewRequest("DELETE", "/api/users/"+userUUID, nil)
		deleteRequest.Header.Set("Authorization", "Bearer "+token)
		deleteResponse := integration.ExecuteRequest(deleteRequest)

		assert.Equal(test, http.StatusNoContent, deleteResponse.Code)

		showRequest, _ := http.NewRequest("GET", "/api/users/"+userUUID, nil)
		showRequest.Header.Set("Authorization", "Bearer "+token)
		showResponse := integration.ExecuteRequest(showRequest)

		assert.Equal(test, http.StatusNotFound, showResponse.Code)
	})

	integration.TestCase(test, "it should fail to create a user with invalid data", func(test *testing.T) {
		token := integration.GetToken()
		payload := map[string]string{
			"name":     "ab", // min=3
			"email":    "not-an-email",
			"password": "short", // min=8
		}
		body, _ := json.Marshal(payload)

		request, _ := http.NewRequest("POST", "/api/users/", bytes.NewBuffer(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+token)

		response := integration.ExecuteRequest(request)

		assert.Equal(test, http.StatusUnprocessableEntity, response.Code)
	})

	integration.TestCase(test, "it should fail to access users without token", func(test *testing.T) {
		request, _ := http.NewRequest("GET", "/api/users/", nil)
		response := integration.ExecuteRequest(request)

		assert.Equal(test, http.StatusUnauthorized, response.Code)
	})
}
