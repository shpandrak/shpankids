package assignment

import (
	"context"
	"shpankids/infra/database/datekvs"
	"shpankids/infra/database/kvstore"
	"shpankids/shpankids"
	"time"
)

type dbUserTaskStatus struct {
	Comment    string                     `json:"comment"`
	Status     shpankids.AssignmentStatus `json:"status"`
	StatusTime time.Time                  `json:"statusTime"`
}

type UserTaskStatusRepository datekvs.DateKvStore[dbUserTaskStatus]

func NewUserTaskStatusRepository(ctx context.Context, kvs kvstore.RawJsonStore, userId string) (UserTaskStatusRepository, error) {
	byUserKvs, err := kvs.CreateSpaceStore(ctx, []string{"users", userId})
	if err != nil {
		return nil, err
	}
	return datekvs.NewDateKvsImpl[dbUserTaskStatus](byUserKvs), nil
}
