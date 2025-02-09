package user

import (
	"shpankids/infra/database/kvstore"
	"time"
)

type UserRepository kvstore.JsonKvStore[string, dbUser]
type dbUser struct {
	Email     string    `json:"email"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	BirthDate time.Time `json:"birthDate"`
}

func NewUserRepository(store kvstore.RawJsonStore) UserRepository {
	return kvstore.NewJsonKvStoreImpl[string, dbUser](
		store,
		"users",
		kvstore.StringKeyToString,
		kvstore.StringToKey,
	)
}
