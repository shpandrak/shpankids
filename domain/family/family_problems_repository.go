package family

import (
	"context"
	"shpankids/infra/database/kvstore"
	"shpankids/shpankids"
	"time"
)

type dbFamilyProblem struct {
	Title        string                     `json:"title"`
	Description  string                     `json:"description"`
	Created      time.Time                  `json:"created"`
	Hints        []string                   `json:"hints"`
	MemberIds    []string                   `json:"memberIds"`
	Alternatives []dbProblemAlternative     `json:"alternatives"`
	Status       shpankids.FamilyTaskStatus `json:"status"`
	StatusDate   time.Time                  `json:"statusDate"`
}

type dbProblemAlternative struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Correct     bool   `json:"correct"`
}

type problemsRepository kvstore.JsonKvStore[string, dbFamilyProblem]

func newFamilyProblemsRepository(ctx context.Context, kvs kvstore.RawJsonStore, familyId string) (problemsRepository, error) {
	familyProblemsStore, err := kvs.CreateSpaceStore(ctx, []string{familiesSpaceStoreUri, familyId})
	if err != nil {
		return nil, err
	}

	return kvstore.NewJsonKvStoreImpl[string, dbFamilyProblem](
		familyProblemsStore,
		"familyProblems",
		kvstore.StringKeyToString,
		kvstore.StringToKey,
	), nil
}
