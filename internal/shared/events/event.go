package events

type Event interface {
	RoutingKey() string
	ToJson() ([]byte, error)
	FromJson(data []byte) error
}
