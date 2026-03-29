package auth

import (
	"auth/internal/domain/auth/services"
	userModel "auth/internal/domain/user/model"
	"auth/tests/integration"
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
)

func TestAuthModule(test *testing.T) {
	integration.TestCase(test, "it should register a user", func(test *testing.T) {
		payload := map[string]string{
			"name":     gofakeit.Name(),
			"email":    gofakeit.Email(),
			"password": "password123",
		}
		body, _ := json.Marshal(payload)

		request, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
		request.Header.Set("Content-Type", "application/json")

		response := integration.ExecuteRequest(request)

		assert.Equal(test, http.StatusCreated, response.Code)

		var data map[string]any
		_ = json.Unmarshal(response.Body.Bytes(), &data)
		assert.Contains(test, data, "access_token")

		cookies := response.Result().Cookies()

		var found bool
		for _, cookie := range cookies {
			if cookie.Name == "refresh_token" {
				found = true
				break
			}
		}
		assert.True(test, found, "refresh_token cookie not found in register")
	})

	integration.TestCase(test, "it should login a user", func(test *testing.T) {
		email := gofakeit.Email()
		password := "password123"

		registerPayload := map[string]string{
			"name":     gofakeit.Name(),
			"email":    email,
			"password": password,
		}
		registerBody, _ := json.Marshal(registerPayload)
		registerRequest, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(registerBody))
		registerRequest.Header.Set("Content-Type", "application/json")
		integration.ExecuteRequest(registerRequest)

		loginPayload := map[string]string{
			"email":    email,
			"password": password,
		}
		loginBody, _ := json.Marshal(loginPayload)
		loginReq, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(loginBody))
		loginReq.Header.Set("Content-Type", "application/json")

		response := integration.ExecuteRequest(loginReq)

		assert.Equal(test, http.StatusOK, response.Code)

		var data map[string]any
		_ = json.Unmarshal(response.Body.Bytes(), &data)

		assert.Contains(test, data, "access_token")

		cookies := response.Result().Cookies()

		var found bool
		for _, cookie := range cookies {
			if cookie.Name == "refresh_token" {
				found = true
				break
			}
		}

		assert.True(test, found, "refresh_token cookie not found")
	})

	integration.TestCase(test, "it should refresh token", func(test *testing.T) {
		email := gofakeit.Email()
		password := "password123"

		registerPayload := map[string]string{
			"name":     gofakeit.Name(),
			"email":    email,
			"password": password,
		}
		registerBody, _ := json.Marshal(registerPayload)
		registerRequest, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(registerBody))
		registerRequest.Header.Set("Content-Type", "application/json")
		registerResponse := integration.ExecuteRequest(registerRequest)

		assert.Equal(test, http.StatusCreated, registerResponse.Code)

		cookies := registerResponse.Result().Cookies()

		var refreshTokenCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "refresh_token" {
				refreshTokenCookie = cookie
				break
			}
		}

		assert.NotNil(test, refreshTokenCookie)

		if refreshTokenCookie != nil {
			refreshReq, _ := http.NewRequest("POST", "/api/auth/refresh", nil)
			refreshReq.AddCookie(refreshTokenCookie)

			response := integration.ExecuteRequest(refreshReq)

			assert.Equal(test, http.StatusOK, response.Code)

			var data map[string]any
			_ = json.Unmarshal(response.Body.Bytes(), &data)

			assert.Contains(test, data, "access_token")
		}
	})

	integration.TestCase(test, "it should logout a user", func(test *testing.T) {
		email := gofakeit.Email()
		password := "password123"

		registerPayload := map[string]string{
			"name":     gofakeit.Name(),
			"email":    email,
			"password": password,
		}
		registerBody, _ := json.Marshal(registerPayload)
		registerRequest, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(registerBody))
		registerRequest.Header.Set("Content-Type", "application/json")
		registerResp := integration.ExecuteRequest(registerRequest)

		assert.Equal(test, http.StatusCreated, registerResp.Code)

		cookies := registerResp.Result().Cookies()

		var refreshTokenCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "refresh_token" {
				refreshTokenCookie = cookie
				break
			}
		}

		assert.NotNil(test, refreshTokenCookie)

		if refreshTokenCookie != nil {
			logoutRequest, _ := http.NewRequest("DELETE", "/api/auth/logout", nil)
			logoutRequest.AddCookie(refreshTokenCookie)

			response := integration.ExecuteRequest(logoutRequest)
			assert.Equal(test, http.StatusNoContent, response.Code)

			refreshRequest, _ := http.NewRequest("POST", "/api/auth/refresh", nil)
			refreshRequest.AddCookie(refreshTokenCookie)
			refreshResponse := integration.ExecuteRequest(refreshRequest)

			assert.Equal(test, http.StatusUnauthorized, refreshResponse.Code)
		}
	})

	integration.TestCase(test, "it should verify a user email", func(test *testing.T) {
		email := gofakeit.Email()
		password := "password123"

		registerPayload := map[string]string{
			"name":     gofakeit.Name(),
			"email":    email,
			"password": password,
		}
		registerBody, _ := json.Marshal(registerPayload)
		registerRequest, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(registerBody))
		registerRequest.Header.Set("Content-Type", "application/json")
		integration.ExecuteRequest(registerRequest)

		// Buscar el usuario para obtener su UUID
		appInstance, err := integration.GetApp()
		assert.NoError(test, err)
		userRepo := userModel.NewRepository(appInstance.Container.DefaultConnection)
		user, _ := userRepo.FindByEmail(email)

		// Generar un token de verificación usando el servicio
		jwtService := services.NewJWTService(appInstance.Config.Auth)
		token, _ := jwtService.GenerateEmailVerificationToken(user.UUID.String())

		// Verificar el email
		verifyPayload := map[string]string{
			"token": token,
		}
		verifyBody, _ := json.Marshal(verifyPayload)
		verifyReq, _ := http.NewRequest("POST", "/api/auth/verify-email", bytes.NewBuffer(verifyBody))
		verifyReq.Header.Set("Content-Type", "application/json")

		response := integration.ExecuteRequest(verifyReq)

		assert.Equal(test, http.StatusOK, response.Code)

		// Verificar en la DB que email_verified_at no sea null
		updatedUser, _ := userRepo.FindByEmail(email)
		assert.NotNil(test, updatedUser.EmailVerifiedAt)
	})
}
