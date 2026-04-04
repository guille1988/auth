package data

type ResendVerification struct {
	Email string `json:"email" binding:"required,email"`
}
