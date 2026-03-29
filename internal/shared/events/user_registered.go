package events

import "encoding/json"

type UserRegistered struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func NewUserRegistered(email, name string) *UserRegistered {
	return &UserRegistered{
		Email: email,
		Name:  name,
	}
}

func (event *UserRegistered) Exchange() string {
	return "user.events"
}

func (event *UserRegistered) RoutingKey() string {
	return "user.registered"
}

func (event *UserRegistered) ToJson() ([]byte, error) {
	return json.Marshal(event)
}

func (event *UserRegistered) FromJson(data []byte) error {
	return json.Unmarshal(data, event)
}
