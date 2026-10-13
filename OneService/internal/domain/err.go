package domain

import "errors"

var (
	ErrExecOutbox   = errors.New("error executing outbox")
	ErrTopicIsEmpty = errors.New("topic is empty")
	ErrEventIsEmpty = errors.New("event is empty")
)
