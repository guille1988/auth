package data

type VerifyEmail struct {
	Token string `json:"token" validate:"required"`
}
