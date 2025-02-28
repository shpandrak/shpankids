package family

import (
	"context"
	"shpankids/infra/database/kvstore"
	"shpankids/shpankids"
	"time"
)

const familiesSpaceStoreUri = "families"

type repository kvstore.JsonKvStore[string, dbFamily]

type dbFamilyMember struct {
	UserId string         `json:"userId"`
	Role   shpankids.Role `json:"role"`
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
		familiesSpaceStoreUri,
		kvstore.StringKeyToString,
		kvstore.StringToKey,
	)
}

func createFamilyUserRootStore(
	ctx context.Context,
	kvs kvstore.RawJsonStore,
	familyId string,
	userId string,
) (kvstore.RawJsonStore, error) {
	familyUserRootRepo, err := kvs.CreateSpaceStore(ctx, []string{
		familiesSpaceStoreUri,
		familyId,
		"users",
		userId,
	})
	return familyUserRootRepo, err
}
