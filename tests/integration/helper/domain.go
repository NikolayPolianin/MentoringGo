package helper

import "time"

type User struct {
	Name      string
	Age       int32
	Birthday  string
	LocalDate time.Duration
}

func UserCreate(name string, age int32, birthday string, localDate time.Duration) User {
	return User{
		Name:      name,
		Age:       age,
		Birthday:  birthday,
		LocalDate: localDate,
	}
}
