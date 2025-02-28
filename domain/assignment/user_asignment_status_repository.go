package assignment

import (
	"context"
	"shpankids/infra/database/datekvs"
	"shpankids/infra/database/kvstore"
	"time"
)

type dbAssignmentStatus struct {
	Comments       []string  `json:"comments,omitempty"`
	PartsCompleted int       `json:"partsCompleted"`
	LastUpdated    time.Time `json:"statusTime"`
}

type assignmentStatusRepo datekvs.DateKvStore[dbAssignmentStatus]

func newAssignmentStatusRepo(ctx context.Context, kvs kvstore.RawJsonStore, userId string) (assignmentStatusRepo, error) {
	byUserKvs, err := kvs.CreateSpaceStore(ctx, []string{"users", userId})
	if err != nil {
		return nil, err
	}
	return datekvs.NewDateKvsImpl[dbAssignmentStatus](byUserKvs), nil
}
