package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"
)

type User struct {
	name string
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_ = retry(ctx, 22, func() (*User, error) {
		user, err := CreateUser("name")
		if err != nil {
			return nil, err
		}
		return user, nil
	})
}

func retry(ctx context.Context, maxAttempt int, fn func() (*User, error)) error {
	afterTime := 300 * time.Millisecond
	AfterErr := errors.New("Retrying false")

	for i := 0; i < maxAttempt; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			_, err := fn()
			if err == nil {
				return nil
			}
			slog.Error("Retrying after error:", err.Error())
		}
		<-time.After(afterTime)
	}
	return AfterErr
}

func NewUser(name string) (*User, error) {
	user := &User{name: name}
	return user, nil
}
func CreateUser(name string) (*User, error) {
	user, err := NewUser(name)
	if err != nil {
		return nil, err
	}
	if err = user.Create(); err != nil {
		return nil, err
	}
	return user, nil
}
func (u *User) Create() error {
	isTrue := func() bool {
		num := rand.Intn(21)
		if num%3 != 0 {
			return false
		}
		return true
	}
	if !isTrue() {
		return errors.New("User Create Failed")
	}
	fmt.Println("User: ", u.name)
	return nil
}
