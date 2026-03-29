package events

type Event interface {
	Exchange() string
	RoutingKey() string
	ToJson() ([]byte, error)
	FromJson(data []byte) error
}
