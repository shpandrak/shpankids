package family

import (
	"context"
	"shpankids/domain/problemset"
	"shpankids/infra/database/kvstore"
)

const problemSetsRepoUri = "problemSets"

type familyProblemsRepository problemset.ProblemsRepository
type familyProblemSetsRepository problemset.ProblemSetsRepository
type familyProblemSolutionsRepository problemset.ProblemSolutionsRepository

func newFamilyProblemsSolutionsRepository(
	ctx context.Context,
	kvs kvstore.RawJsonStore,
	familyId string,
	userId string,
	problemSetId string,
) (familyProblemSolutionsRepository, error) {
	rootFamilyStore, err := createFamilyUserRootRepo(ctx, kvs, familyId, userId)
	if err != nil {
		return nil, err
	}

	return problemset.NewProblemsSolutionsRepository(ctx, rootFamilyStore, problemSetId)
}

func newFamilyProblemsRepository(
	ctx context.Context,
	kvs kvstore.RawJsonStore,
	familyId string,
	userId string,
	problemSetId string,
) (familyProblemsRepository, error) {

	familyProblemsStore, err := createFamilyUserRootRepo(ctx, kvs, familyId, userId)
	if err != nil {
		return nil, err
	}

	return problemset.NewProblemSetProblemsRepository(ctx, familyProblemsStore, problemSetId)
}

func newProblemSetsRepository(
	ctx context.Context,
	kvs kvstore.RawJsonStore,
	familyId string,
	userId string,
) (familyProblemSetsRepository, error) {
	familyUserRootRepo, err := createFamilyUserRootRepo(ctx, kvs, familyId, userId)
	if err != nil {
		return nil, err
	}

	return problemset.NewProblemSetsRepository(familyUserRootRepo)
}

func createFamilyUserRootRepo(
	ctx context.Context,
	kvs kvstore.RawJsonStore,
	familyId string,
	userId string,
) (kvstore.RawJsonStore, error) {
	familyUserRootRepo, err := kvs.CreateSpaceStore(ctx, []string{
		familiesSpaceStoreUri,
		familyId,
		"users",
		userId,
	})
	return familyUserRootRepo, err
}
