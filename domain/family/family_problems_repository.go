package family

import (
	"context"
	"shpankids/infra/database/archkvs"
	"shpankids/infra/database/kvstore"
	"shpankids/shpankids"
	"time"
)

const problemSetsRepoUri = "problemSets"

type problemsRepository archkvs.ArchivedKvs[string, dbFamilyProblem]

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

	return archkvs.NewArchivedKvsImpl[string, dbFamilyProblem](
		ctx,
		familyProblemsStore,
		"problems",
		kvstore.StringKeyToString,
		kvstore.StringToKey,
	)
}

type problemSetsRepository kvstore.JsonKvStore[string, dbFamilyProblemSet]

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

	return kvstore.NewJsonKvStoreImpl[string, dbFamilyProblemSet](
		familyProblemsStore,
		problemSetsRepoUri,
		kvstore.StringKeyToString,
		kvstore.StringToKey,
	), nil
}

type dbFamilyProblemSet struct {
	Title       string                           `json:"title"`
	Description string                           `json:"description"`
	Created     time.Time                        `json:"created"`
	Status      shpankids.FamilyAssignmentStatus `json:"status"`
	StatusDate  time.Time                        `json:"statusDate"`
}

type dbFamilyProblem struct {
	Title        string                 `json:"title"`
	Description  string                 `json:"description"`
	Created      time.Time              `json:"created"`
	Hints        []string               `json:"hints,omitempty"`
	Explanation  string                 `json:"explanation,omitempty"`
	Alternatives []dbProblemAlternative `json:"alternatives"`
}

type dbProblemAlternative struct {
	Id          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Correct     bool   `json:"correct"`
}
