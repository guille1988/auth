package dtos

type WelcomeEmail struct {
	Email           string `json:"email"`
	Name            string `json:"name"`
	VerificationURL string `json:"verification_url"`
}
