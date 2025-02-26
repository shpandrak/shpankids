package problemset

import (
	"context"
	"shpankids/infra/database/archkvs"
	"shpankids/infra/database/datekvs"
	"shpankids/infra/database/kvstore"
	"shpankids/shpankids"
	"time"
)

const problemSetsRepoUri = "problemSets"

type ProblemsRepository archkvs.ArchivedKvs[string, DbProblem]
type ProblemSetsRepository kvstore.JsonKvStore[string, DbProblemSet]
type ProblemSolutionsRepository datekvs.DateKvStore[DbProblemSolution]

func NewProblemSetProblemsRepository(
	ctx context.Context,
	kvs kvstore.RawJsonStore,
	problemSetId string,
) (ProblemsRepository, error) {
	psStore, err := createRootProblemSetStore(ctx, kvs, problemSetId)
	if err != nil {
		return nil, err
	}

	return archkvs.NewArchivedKvsImpl[string, DbProblem](
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

func NewProblemsSolutionsRepository(
	ctx context.Context,
	kvs kvstore.RawJsonStore,
	problemSetId string,
) (ProblemSolutionsRepository, error) {
	psStore, err := createRootProblemSetStore(ctx, kvs, problemSetId)
	if err != nil {
		return nil, err
	}

	return datekvs.NewDateKvsImpl[DbProblemSolution](psStore), nil
}

func NewProblemSetsRepository(
	kvs kvstore.RawJsonStore,
) (ProblemSetsRepository, error) {
	return kvstore.NewJsonKvStoreImpl[string, DbProblemSet](
		kvs,
		problemSetsRepoUri,
		kvstore.StringKeyToString,
		kvstore.StringToKey,
	), nil
}

type DbProblemSet struct {
	Title       string                           `json:"title"`
	Description string                           `json:"description"`
	Created     time.Time                        `json:"created"`
	Status      shpankids.FamilyAssignmentStatus `json:"status"`
	StatusDate  time.Time                        `json:"statusDate"`
}

type DbProblem struct {
	Title       string                     `json:"title"`
	Description string                     `json:"description"`
	Created     time.Time                  `json:"created"`
	Hints       []string                   `json:"hints,omitempty"`
	Explanation string                     `json:"explanation,omitempty"`
	Answers     map[string]DbProblemAnswer `json:"answers"`
}

type DbProblemAnswer struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Correct     bool   `json:"correct,omitempty"`
}

type DbProblemSolution struct {
	SelectedAnswerId string `json:"selectedAnswerId"`
	Correct          bool   `json:"correct"`
}
