package problemset

import (
	"context"
	"shpankids/infra/database/archkvs"
	"shpankids/infra/database/kvstore"
	"shpankids/shpankids"
	"time"
)

const problemSetsRepoUri = "problemSets"

type problemsRepository kvstore.JsonKvStore[string, DbProblem]
type problemSetsRepository kvstore.JsonKvStore[string, DbProblemSet]

func newProblemSetProblemsRepository(
	ctx context.Context,
	kvs kvstore.RawJsonStore,
	problemSetId string,
) (problemsRepository, error) {
	psPStore, err := kvs.CreateSpaceStore(ctx, []string{
		problemSetsRepoUri,
		problemSetId,
	})
	if err != nil {
		return nil, err
	}

	return archkvs.NewArchivedKvsImpl[string, DbProblem](
		ctx,
		psPStore,
		"problems",
		kvstore.StringKeyToString,
		kvstore.StringToKey,
	)
}

func newProblemSetsRepository(
	kvs kvstore.RawJsonStore,
) (problemSetsRepository, error) {
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

type dbProblemSolution struct {
	SelectedAnswerId string `json:"selectedAnswerId"`
	Correct          bool   `json:"correct"`
}
