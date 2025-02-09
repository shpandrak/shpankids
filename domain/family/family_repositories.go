package family

import (
	"shpankids/infra/database/kvstore"
	"time"
)

type repository kvstore.JsonKvStore[string, dbFamily]

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

type dbFamilyMember struct {
	UserId string `json:"userId"`
	Role   Role   `json:"role"`
}

type dbFamily struct {
	Id        string           `json:"id"`
	Name      string           `json:"name"`
	CreatedBy string           `json:"createdBy"`
	CreatedAt time.Time        `json:"createdAt"`
	Members   []dbFamilyMember `json:"members"`
}

func newFamilyRepository(store kvstore.RawJsonStore) repository {
	return kvstore.NewJsonKvStoreImpl[string, dbFamily](
		store,
		"families",
		kvstore.StringKeyToString,
		kvstore.StringToKey,
	)
}
