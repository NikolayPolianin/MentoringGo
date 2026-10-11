package domain

import (
	"errors"
	"github.com/google/uuid"
	"log/slog"
	"time"
)

type User struct {
	UuID      uuid.UUID `json:"-"`
	Id        int64     `json:"id,omitempty"`
	Name      string    `json:"name"`
	Age       int       `json:"age"`
	Birthday  string    `json:"birthday"`
	LocalDate time.Time `json:"localDate"`
}

func NewUser(name string, age int32, birthDay string) (*User, error) {
	user := &User{
		Name:      name,
		Age:       int(age),
		Birthday:  birthDay,
		LocalDate: time.Now(),
	}

	err := ValidateUser(user)
	if err != nil {
		slog.Warn("Validation error:", "errors:", err)
		return nil, err
	}
	return user, nil
}

func ValidateUser(user *User) error {
	var errs []error
	if user.Name == "" {
		errs = append(errs, errors.New("name is required"))
	}
	if user.Age < 18 {
		errs = append(errs, errors.New("age is less than 18"))
	}
	return errors.Join(errs...)
}
