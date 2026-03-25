package events

import "encoding/json"

type UserRegistered struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func NewUserRegisteredEvent(email, name string) *UserRegistered {
	return &UserRegistered{
		Email: email,
		Name:  name,
	}
}

func (event *UserRegistered) GetRoutingKey() string {
	return "user.registered"
}

func (event *UserRegistered) ToJson() ([]byte, error) {
	return json.Marshal(event)
}
