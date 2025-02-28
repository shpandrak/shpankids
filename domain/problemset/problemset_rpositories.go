package problemset

import (
	"context"
	"shpankids/infra/database/archkvs"
	"shpankids/infra/database/datekvs"
	"shpankids/infra/database/kvstore"
	"time"
)

const problemSetsRepoUri = "problemSets"

type problemsRepository archkvs.ArchivedKvs[string, dbProblem]
type problemSetsRepository kvstore.JsonKvStore[string, dbProblemSet]
type problemSolutionsRepository datekvs.DateKvStore[dbProblemSolution]

func newProblemSetProblemsRepository(
	ctx context.Context,
	kvs kvstore.RawJsonStore,
	problemSetId string,
) (problemsRepository, error) {
	psStore, err := createRootProblemSetStore(ctx, kvs, problemSetId)
	if err != nil {
		return nil, err
	}

	return archkvs.NewArchivedKvsImpl[string, dbProblem](
		ctx,
		psStore,
		"problems",
		kvstore.StringKeyToString,
		kvstore.StringToKey,
	)
}

func createRootProblemSetStore(
	ctx context.Context,
	kvs kvstore.RawJsonStore,
	problemSetId string,
) (kvstore.RawJsonStore, error) {
	return kvs.CreateSpaceStore(ctx, []string{
		problemSetsRepoUri,
		problemSetId,
	})
}

func newProblemsSolutionsRepository(
	ctx context.Context,
	kvs kvstore.RawJsonStore,
	problemSetId string,
) (problemSolutionsRepository, error) {
	psStore, err := createRootProblemSetStore(ctx, kvs, problemSetId)
	if err != nil {
		return nil, err
	}

	return datekvs.NewDateKvsImpl[dbProblemSolution](psStore), nil
}

func newProblemSetsRepository(
	kvs kvstore.RawJsonStore,
) (problemSetsRepository, error) {
	return kvstore.NewJsonKvStoreImpl[string, dbProblemSet](
		kvs,
		problemSetsRepoUri,
		kvstore.StringKeyToString,
		kvstore.StringToKey,
	), nil
}

type dbProblemSet struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Created     time.Time `json:"created"`
	//Status      shpankids.FamilyAssignmentStatus `json:"status"`
	//StatusDate  time.Time                        `json:"statusDate"`
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
