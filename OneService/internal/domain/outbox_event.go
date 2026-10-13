package domain

type Event struct {
	Topic string
	Key   []byte
	Value []byte
}
