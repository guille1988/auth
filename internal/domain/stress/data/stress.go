package data

type Stress struct {
	Email string `json:"email" binding:"required,email"`
	Name  string `json:"name"  binding:"required,min=1,max=100"`
}
