package user

import (
	"context"
	"fmt"
	"shpankids/infra/database/kvstore"
	"shpankids/internal/infra/util"
	"shpankids/shpankids"
	"time"
)

type manager struct {
	repository UserRepository
}

func (um *manager) DeleteUser(ctx context.Context, email string) error {
	err := um.repository.Unset(ctx, email)
	if err != nil {
		return err
	}
	return nil
}

func NewUserManager(kvs kvstore.RawJsonStore) shpankids.UserManager {

	return &manager{
		repository: NewUserRepository(kvs),
	}
}

func (um *manager) CreateUser(ctx context.Context, email string, firstName string, lastNae string, birthDate time.Time) error {
	// Check if user already exists
	fnd, err := um.FindUser(ctx, email)
	if err != nil {
		return err
	}
	if fnd != nil {
		return util.DuplicateInputError(fmt.Errorf("user with email %s already exists", email))
	}

	// Create user
	dbUser := dbUser{
		Email:     email,
		FirstName: firstName,
		LastName:  lastNae,
		BirthDate: birthDate,
	}
	err = um.repository.Set(ctx, email, dbUser)
	if err != nil {
		return err
	}
	return nil

}

func (um *manager) GetUser(ctx context.Context, email string) (*shpankids.User, error) {
	dbUsr, err := um.repository.Get(ctx, email)
	if err != nil {
		return nil, err
	}
	return mapUser(dbUsr), nil

}

func (um *manager) FindUser(ctx context.Context, email string) (*shpankids.User, error) {
	dbUsr, err := um.repository.Find(ctx, email)
	if err != nil {
		return nil, err
	}
	if dbUsr == nil {
		return nil, nil
	}
	return mapUser(*dbUsr), nil
}

func mapUser(dbUser dbUser) *shpankids.User {
	return &shpankids.User{
		Email:     dbUser.Email,
		FirstName: dbUser.FirstName,
		LastName:  dbUser.LastName,
		BirthDate: dbUser.BirthDate,
	}
}
