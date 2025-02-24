package family

import (
	"context"
	"shpankids/domain/problemset"
	"shpankids/infra/database/archkvs"
	"shpankids/infra/database/datekvs"
	"shpankids/infra/database/kvstore"
)

const problemSetsRepoUri = "problemSets"

type familyProblemsRepository archkvs.ArchivedKvs[string, problemset.DbProblem]
type familyProblemSetsRepository kvstore.JsonKvStore[string, problemset.DbProblemSet]
type familyProblemSolutionsRepository datekvs.DateKvStore[dbProblemSolution]

func newFamilyProblemsSolutionsRepository(
	ctx context.Context,
	kvs kvstore.RawJsonStore,
	familyId string,
	userId string,
	problemSetId string,
) (familyProblemSolutionsRepository, error) {
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

func newFamilyProblemsRepository(ctx context.Context, kvs kvstore.RawJsonStore, familyId string, userId string, problemSetId string) (familyProblemsRepository, error) {
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

	return archkvs.NewArchivedKvsImpl[string, problemset.DbProblem](
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
) (familyProblemSetsRepository, error) {
	familyProblemsStore, err := kvs.CreateSpaceStore(ctx, []string{
		familiesSpaceStoreUri,
		familyId,
		"users",
		userId,
	})
	if err != nil {
		return nil, err
	}

	return kvstore.NewJsonKvStoreImpl[string, problemset.DbProblemSet](
		familyProblemsStore,
		problemSetsRepoUri,
		kvstore.StringKeyToString,
		kvstore.StringToKey,
	), nil
}

type dbProblemSolution struct {
	SelectedAnswerId string `json:"selectedAnswerId"`
	Correct          bool   `json:"correct"`
}
