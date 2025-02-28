package assignment

import (
	"context"
	"shpankids/infra/database/archkvs"
	"shpankids/infra/database/kvstore"
	"shpankids/shpankids"
	"time"
)

type dbUserAssignment struct {
	AssignmentType  shpankids.AssignmentType `json:"assignmentType"`
	Title           string                   `json:"title"`
	Created         time.Time                `json:"created"`
	NumberOfParts   int                      `json:"numberOfParts"`
	CreatedByUserId string                   `json:"createdByUserId"`
	Description     string                   `json:"description,omitempty"`
}

type userAssignmentRepository archkvs.ArchivedKvs[string, dbUserAssignment]

func newUserAssignmentsRepository(
	ctx context.Context,
	kvs kvstore.RawJsonStore,
	userId string) (userAssignmentRepository, error) {
	byUserKvs, err := kvs.CreateSpaceStore(ctx, []string{"users", userId})
	if err != nil {
		return nil, err
	}
	return archkvs.NewArchivedKvsImpl[string, dbUserAssignment](
		ctx,
		byUserKvs,
		"assignments",
		kvstore.StringKeyToString,
		kvstore.StringToKey,
	)
}
