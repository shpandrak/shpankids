package session

import (
	"shpankids/infra/database/kvstore"
)

type sessionRepository kvstore.JsonKvStore[string, dbSession]
type dbSession struct {
	FamilyId string `json:"familyId"`
	TimeZone string `json:"timeZone"`
}

func newSessionRepository(store kvstore.RawJsonStore) sessionRepository {
	return kvstore.NewJsonKvStoreImpl[string, dbSession](
		store,
		"sessions",
		kvstore.StringKeyToString,
		kvstore.StringToKey,
	)
}
