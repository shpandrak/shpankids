package family

import (
	"context"
	"shpankids/infra/database/archkvs"
	"shpankids/infra/database/datekvs"
	"shpankids/infra/database/kvstore"
	"shpankids/shpankids"
	"time"
)

const problemSetsRepoUri = "problemSets"

type problemsRepository archkvs.ArchivedKvs[string, dbProblem]
type problemSetsRepository kvstore.JsonKvStore[string, dbProblemSet]
type problemSolutionsRepository datekvs.DateKvStore[dbProblemSolution]

func newFamilyProblemsSolutionsRepository(
	ctx context.Context,
	kvs kvstore.RawJsonStore,
	familyId string,
	userId string,
	problemSetId string,
) (problemSolutionsRepository, error) {
	familyProblemsStore, err := kvs.CreateSpaceStore(ctx, []string{
		familiesSpaceStoreUri,
		familyId,
		"users",
		userId,
		problemSetsRepoUri,
		problemSetId,
	})
	if err != nil {
		return nil, err
	}

	return datekvs.NewDateKvsImpl[dbProblemSolution](familyProblemsStore), nil
}

func newFamilyProblemsRepository(ctx context.Context, kvs kvstore.RawJsonStore, familyId string, userId string, problemSetId string) (problemsRepository, error) {
	familyProblemsStore, err := kvs.CreateSpaceStore(ctx, []string{
		familiesSpaceStoreUri,
		familyId,
		"users",
		userId,
		problemSetsRepoUri,
		problemSetId,
	})
	if err != nil {
		return nil, err
	}

	return archkvs.NewArchivedKvsImpl[string, dbProblem](
		ctx,
		familyProblemsStore,
		"problems",
		kvstore.StringKeyToString,
		kvstore.StringToKey,
	)
}

func newProblemSetsRepository(
	ctx context.Context,
	kvs kvstore.RawJsonStore,
	familyId string,
	userId string,
) (problemSetsRepository, error) {
	familyProblemsStore, err := kvs.CreateSpaceStore(ctx, []string{
		familiesSpaceStoreUri,
		familyId,
		"users",
		userId,
	})
	if err != nil {
		return nil, err
	}

	return kvstore.NewJsonKvStoreImpl[string, dbProblemSet](
		familyProblemsStore,
		problemSetsRepoUri,
		kvstore.StringKeyToString,
		kvstore.StringToKey,
	), nil
}

type dbProblemSet struct {
	Title       string                           `json:"title"`
	Description string                           `json:"description"`
	Created     time.Time                        `json:"created"`
	Status      shpankids.FamilyAssignmentStatus `json:"status"`
	StatusDate  time.Time                        `json:"statusDate"`
}

type dbProblem struct {
	Title       string                     `json:"title"`
	Description string                     `json:"description"`
	Created     time.Time                  `json:"created"`
	Hints       []string                   `json:"hints,omitempty"`
	Explanation string                     `json:"explanation,omitempty"`
	Answers     map[string]dbProblemAnswer `json:"answers"`
}

type dbProblemAnswer struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Correct     bool   `json:"correct,omitempty"`
}

type dbProblemSolution struct {
	SelectedAnswerId string `json:"selectedAnswerId"`
	Correct          bool   `json:"correct"`
}
