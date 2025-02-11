package family

import (
	"context"
	"shpankids/infra/database/kvstore"
	"shpankids/shpankids"
	"time"
)

type dbFamilyTask struct {
	Title       string                     `json:"title"`
	Description string                     `json:"description"`
	MemberIds   []string                   `json:"memberIds"`
	Status      shpankids.FamilyTaskStatus `json:"status"`
	StatusDate  time.Time                  `json:"statusDate"`
}

type familyTaskRepository kvstore.JsonKvStore[string, dbFamilyTask]

func newFamilyTaskRepository(ctx context.Context, kvs kvstore.RawJsonStore, familyId string) (familyTaskRepository, error) {
	familyTasksStore, err := kvs.CreateSpaceStore(ctx, []string{"families", familyId})
	if err != nil {
		return nil, err
	}

	return kvstore.NewJsonKvStoreImpl[string, dbFamilyTask](
		familyTasksStore,
		"familyTasks",
		kvstore.StringKeyToString,
		kvstore.StringToKey,
	), nil
}
