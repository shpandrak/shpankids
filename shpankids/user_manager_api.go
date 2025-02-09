package shpankids

import (
	"context"
	"time"
)

type User struct {
	Email     string
	FirstName string
	LastName  string
	BirthDate time.Time
}

type UserManager interface {
	CreateUser(ctx context.Context, email string, firstName string, lastNae string, birthDate time.Time) error
	DeleteUser(ctx context.Context, email string) error
	GetUser(ctx context.Context, email string) (*User, error)
	FindUser(ctx context.Context, email string) (*User, error)
}
