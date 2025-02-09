package family

import "shpankids/infra/database/kvstore"

type dbFamilyTask struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	MemberIds   []string `json:"memberIds"`
}

type familyTaskRepository kvstore.JsonKvStore[string, dbFamilyTask]

func newFamilyTaskRepository(store kvstore.RawJsonStore) familyTaskRepository {
	return kvstore.NewJsonKvStoreImpl[string, dbFamilyTask](
		store,
		"familyTasks",
		kvstore.StringKeyToString,
		kvstore.StringToKey,
	)
}
