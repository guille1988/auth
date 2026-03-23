package data

type RefreshToken struct {
	UserID uint   `json:"user_id"`
	Device string `json:"device"`
}
