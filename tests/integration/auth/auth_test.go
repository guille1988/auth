package auth

import (
	"auth/internal/domain/auth/services"
	userModel "auth/internal/domain/user/model"
	"auth/tests/integration"
	"bytes"
	"encoding/json"
	"net/http"
	"sync"
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

		appInstance, err := integration.GetApp()
		assert.NoError(test, err)

		userRepository := userModel.NewRepository(appInstance.Container.DefaultConnection)
		user, _ := userRepository.FindByEmail(email)
		jwtService := services.NewJWTService(appInstance.Config.Auth)
		verificationToken, _ := jwtService.GenerateEmailVerificationToken(user.UUID.String())

		verifyPayload := map[string]string{"token": verificationToken}
		verifyBody, _ := json.Marshal(verifyPayload)
		verifyRequest, _ := http.NewRequest("POST", "/api/auth/verify-email", bytes.NewBuffer(verifyBody))
		verifyRequest.Header.Set("Content-Type", "application/json")
		integration.ExecuteRequest(verifyRequest)

		loginPayload := map[string]string{
			"email":    email,
			"password": password,
		}
		loginBody, _ := json.Marshal(loginPayload)
		loginRequest, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(loginBody))
		loginRequest.Header.Set("Content-Type", "application/json")

		response := integration.ExecuteRequest(loginRequest)

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

		updatedUser, _ := userRepository.FindByEmail(email)
		assert.NotNil(test, updatedUser.LastLoginAt)
	})

	integration.TestCase(test, "it should return 403 when logging in with unverified email", func(test *testing.T) {
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
		loginRequest, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(loginBody))
		loginRequest.Header.Set("Content-Type", "application/json")

		response := integration.ExecuteRequest(loginRequest)

		assert.Equal(test, http.StatusForbidden, response.Code)
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
			refreshRequest, _ := http.NewRequest("POST", "/api/auth/refresh", nil)
			refreshRequest.AddCookie(refreshTokenCookie)

			response := integration.ExecuteRequest(refreshRequest)

			assert.Equal(test, http.StatusOK, response.Code)

			var data map[string]any
			_ = json.Unmarshal(response.Body.Bytes(), &data)

			assert.Contains(test, data, "access_token")
		}
	})

	integration.TestCase(test, "it should only let one of two concurrent refresh requests succeed with the same token", func(test *testing.T) {
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

		var refreshTokenCookie *http.Cookie
		for _, cookie := range registerResponse.Result().Cookies() {
			if cookie.Name == "refresh_token" {
				refreshTokenCookie = cookie
				break
			}
		}
		assert.NotNil(test, refreshTokenCookie)

		const concurrentAttempts = 10
		statusCodes := make(chan int, concurrentAttempts)
		var waitGroup sync.WaitGroup

		for i := 0; i < concurrentAttempts; i++ {
			waitGroup.Add(1)
			go func() {
				defer waitGroup.Done()
				refreshRequest, _ := http.NewRequest("POST", "/api/auth/refresh", nil)
				refreshRequest.AddCookie(refreshTokenCookie)
				response := integration.ExecuteRequest(refreshRequest)
				statusCodes <- response.Code
			}()
		}

		waitGroup.Wait()
		close(statusCodes)

		successCount := 0
		unauthorizedCount := 0
		for code := range statusCodes {
			switch code {
			case http.StatusOK:
				successCount++
			case http.StatusUnauthorized:
				unauthorizedCount++
			}
		}

		assert.Equal(test, 1, successCount, "exactly one concurrent refresh with the same token must succeed")
		assert.Equal(test, concurrentAttempts-1, unauthorizedCount, "every other concurrent refresh with the same (now consumed) token must be rejected")
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

	integration.TestCase(test, "it should return forbidden status code when email is not verified", func(test *testing.T) {
		payload := map[string]string{
			"name":     gofakeit.Name(),
			"email":    gofakeit.Email(),
			"password": "password123",
		}
		body, _ := json.Marshal(payload)

		registerRequest, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
		registerRequest.Header.Set("Content-Type", "application/json")
		registerResponse := integration.ExecuteRequest(registerRequest)

		var data map[string]any
		_ = json.Unmarshal(registerResponse.Body.Bytes(), &data)
		accessToken := data["access_token"].(string)

		validateRequest, _ := http.NewRequest("GET", "/api/auth/validate", nil)
		validateRequest.Header.Set("Authorization", "Bearer "+accessToken)

		response := integration.ExecuteRequest(validateRequest)

		assert.Equal(test, http.StatusForbidden, response.Code)
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

		appInstance, err := integration.GetApp()
		assert.NoError(test, err)
		userRepo := userModel.NewRepository(appInstance.Container.DefaultConnection)
		user, _ := userRepo.FindByEmail(email)

		jwtService := services.NewJWTService(appInstance.Config.Auth)
		token, _ := jwtService.GenerateEmailVerificationToken(user.UUID.String())

		verifyPayload := map[string]string{
			"token": token,
		}
		verifyBody, _ := json.Marshal(verifyPayload)
		verifyRequest, _ := http.NewRequest("POST", "/api/auth/verify-email", bytes.NewBuffer(verifyBody))
		verifyRequest.Header.Set("Content-Type", "application/json")

		response := integration.ExecuteRequest(verifyRequest)

		assert.Equal(test, http.StatusNoContent, response.Code)

		updatedUser, _ := userRepo.FindByEmail(email)
		assert.NotNil(test, updatedUser.EmailVerifiedAt)
	})

	integration.TestCase(test, "it should resend verification email", func(test *testing.T) {
		email := gofakeit.Email()

		registerPayload := map[string]string{
			"name":     gofakeit.Name(),
			"email":    email,
			"password": "password123",
		}
		registerBody, _ := json.Marshal(registerPayload)
		registerRequest, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(registerBody))
		registerRequest.Header.Set("Content-Type", "application/json")
		integration.ExecuteRequest(registerRequest)

		resendPayload := map[string]string{"email": email}
		resendBody, _ := json.Marshal(resendPayload)
		resendRequest, _ := http.NewRequest("POST", "/api/auth/resend-verification", bytes.NewBuffer(resendBody))
		resendRequest.Header.Set("Content-Type", "application/json")

		response := integration.ExecuteRequest(resendRequest)

		assert.Equal(test, http.StatusNoContent, response.Code)

		appInstance, err := integration.GetApp()
		assert.NoError(test, err)
		userRepo := userModel.NewRepository(appInstance.Container.DefaultConnection)
		user, _ := userRepo.FindByEmail(email)

		jwtService := services.NewJWTService(appInstance.Config.Auth)
		newToken, _ := jwtService.GenerateEmailVerificationToken(user.UUID.String())

		verifyPayload := map[string]string{"token": newToken}
		verifyBody, _ := json.Marshal(verifyPayload)
		verifyRequest, _ := http.NewRequest("POST", "/api/auth/verify-email", bytes.NewBuffer(verifyBody))
		verifyRequest.Header.Set("Content-Type", "application/json")
		integration.ExecuteRequest(verifyRequest)

		updatedUser, _ := userRepo.FindByEmail(email)
		assert.NotNil(test, updatedUser.EmailVerifiedAt)
	})

	integration.TestCase(test, "it should succeed silently when resending for already verified email", func(test *testing.T) {
		email := gofakeit.Email()

		registerPayload := map[string]string{
			"name":     gofakeit.Name(),
			"email":    email,
			"password": "password123",
		}
		registerBody, _ := json.Marshal(registerPayload)
		registerRequest, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(registerBody))
		registerRequest.Header.Set("Content-Type", "application/json")
		integration.ExecuteRequest(registerRequest)

		appInstance, err := integration.GetApp()
		assert.NoError(test, err)
		userRepo := userModel.NewRepository(appInstance.Container.DefaultConnection)
		user, _ := userRepo.FindByEmail(email)

		jwtService := services.NewJWTService(appInstance.Config.Auth)
		token, _ := jwtService.GenerateEmailVerificationToken(user.UUID.String())

		verifyPayload := map[string]string{"token": token}
		verifyBody, _ := json.Marshal(verifyPayload)
		verifyRequest, _ := http.NewRequest("POST", "/api/auth/verify-email", bytes.NewBuffer(verifyBody))
		verifyRequest.Header.Set("Content-Type", "application/json")
		integration.ExecuteRequest(verifyRequest)

		resendPayload := map[string]string{"email": email}
		resendBody, _ := json.Marshal(resendPayload)
		resendRequest, _ := http.NewRequest("POST", "/api/auth/resend-verification", bytes.NewBuffer(resendBody))
		resendRequest.Header.Set("Content-Type", "application/json")

		response := integration.ExecuteRequest(resendRequest)

		assert.Equal(test, http.StatusNoContent, response.Code)
	})

	integration.TestCase(test, "it should succeed silently when resending for non-existent email", func(test *testing.T) {
		resendPayload := map[string]string{"email": gofakeit.Email()}
		resendBody, _ := json.Marshal(resendPayload)
		resendRequest, _ := http.NewRequest("POST", "/api/auth/resend-verification", bytes.NewBuffer(resendBody))
		resendRequest.Header.Set("Content-Type", "application/json")

		response := integration.ExecuteRequest(resendRequest)

		assert.Equal(test, http.StatusNoContent, response.Code)
	})

	integration.TestCase(test, "it should reject an access token used as an email verification token", func(test *testing.T) {
		payload := map[string]string{
			"name":     gofakeit.Name(),
			"email":    gofakeit.Email(),
			"password": "password123",
		}
		body, _ := json.Marshal(payload)

		registerRequest, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
		registerRequest.Header.Set("Content-Type", "application/json")
		registerResponse := integration.ExecuteRequest(registerRequest)

		var data map[string]any
		_ = json.Unmarshal(registerResponse.Body.Bytes(), &data)
		accessToken := data["access_token"].(string)

		verifyPayload := map[string]string{"token": accessToken}
		verifyBody, _ := json.Marshal(verifyPayload)
		verifyRequest, _ := http.NewRequest("POST", "/api/auth/verify-email", bytes.NewBuffer(verifyBody))
		verifyRequest.Header.Set("Content-Type", "application/json")

		response := integration.ExecuteRequest(verifyRequest)

		assert.Equal(test, http.StatusUnprocessableEntity, response.Code)
	})

	integration.TestCase(test, "it should reject an email verification token used as a bearer access token", func(test *testing.T) {
		email := gofakeit.Email()

		registerPayload := map[string]string{
			"name":     gofakeit.Name(),
			"email":    email,
			"password": "password123",
		}
		registerBody, _ := json.Marshal(registerPayload)
		registerRequest, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(registerBody))
		registerRequest.Header.Set("Content-Type", "application/json")
		integration.ExecuteRequest(registerRequest)

		appInstance, err := integration.GetApp()
		assert.NoError(test, err)
		userRepo := userModel.NewRepository(appInstance.Container.DefaultConnection)
		user, _ := userRepo.FindByEmail(email)

		jwtService := services.NewJWTService(appInstance.Config.Auth)
		verificationToken, _ := jwtService.GenerateEmailVerificationToken(user.UUID.String())

		validateRequest, _ := http.NewRequest("GET", "/api/auth/validate", nil)
		validateRequest.Header.Set("Authorization", "Bearer "+verificationToken)

		response := integration.ExecuteRequest(validateRequest)

		assert.Equal(test, http.StatusUnauthorized, response.Code)
	})

	integration.TestCase(test, "it should reject an empty verification token at the validation layer, not the JWT parser", func(test *testing.T) {
		verifyPayload := map[string]string{"token": ""}
		verifyBody, _ := json.Marshal(verifyPayload)
		verifyRequest, _ := http.NewRequest("POST", "/api/auth/verify-email", bytes.NewBuffer(verifyBody))
		verifyRequest.Header.Set("Content-Type", "application/json")

		response := integration.ExecuteRequest(verifyRequest)

		assert.Equal(test, http.StatusUnprocessableEntity, response.Code)

		var data map[string]any
		_ = json.Unmarshal(response.Body.Bytes(), &data)
		assert.Contains(test, data["error"], "Token", "expected a field validation error mentioning the Token field, not a JWT parse error")
		assert.NotContains(test, data["error"], "invalid or expired verification token")
	})
}
