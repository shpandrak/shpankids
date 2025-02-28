package problemset

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"shpankids/domain/ai"
	"shpankids/infra/database/datekvs"
	"shpankids/infra/database/kvstore"
	"shpankids/infra/shpanstream"
	"shpankids/infra/util/functional"
	"shpankids/internal/api"
	"shpankids/internal/infra/util"
	"shpankids/openapi"
	"shpankids/shpankids"
	"time"
)

type Manager struct {
	kvs kvstore.RawJsonStore
}

func NewProblemSetManager(kvs kvstore.RawJsonStore) *Manager {
	return &Manager{kvs: kvs}
}

func (psm *Manager) CreateProblemsInSet(ctx context.Context, problemSetId string, args []shpankids.CreateProblemDto) error {

	repo, err := newProblemSetProblemsRepository(ctx, psm.kvs, problemSetId)
	if err != nil {
		return err
	}
	createdTime := time.Now()
	for _, p := range args {
		if p.Title == "" {
			return util.BadInputError(fmt.Errorf("title is required"))
		}

		if functional.CountSliceNoErr(p.Answers, func(a shpankids.CreateProblemAnswerDto) bool {
			return a.Correct
		}) != 1 {
			return util.BadInputError(fmt.Errorf("one and only one correct answer is required for problem %s", p.Title))
		}

		for _, a := range p.Answers {
			if a.Title == "" {
				return util.BadInputError(fmt.Errorf("alternative title is required"))
			}
		}

		dbAnswers := make(map[string]dbProblemAnswer, len(p.Answers))
		for idx, a := range p.Answers {
			dbAnswers[fmt.Sprintf("%d", idx)] = dbProblemAnswer{
				Title:       a.Title,
				Description: a.Description,
				Correct:     a.Correct,
			}
		}

		// Create the family task in repo
		err = repo.Set(ctx, uuid.NewString(), dbProblem{
			Title:       p.Title,
			Description: p.Description,
			Created:     createdTime,
			Hints:       p.Hints,
			Explanation: p.Explanation,
			Answers:     dbAnswers,
		})
		if err != nil {
			return err
		}
	}
	return nil

}

func (psm *Manager) CreateProblemSet(ctx context.Context, args shpankids.CreateProblemSetDto) error {
	psRepo, err := newProblemSetsRepository(psm.kvs)
	if err != nil {
		return err
	}
	createTime := time.Now()
	// Create the family task in repo
	return psRepo.Set(
		ctx,
		args.ProblemSetId,
		dbProblemSet{
			Title:       args.Title,
			Description: args.Description,
			Created:     createTime,
		})
}

func (psm *Manager) ListProblemSets(ctx context.Context) shpanstream.Stream[shpankids.ProblemSetDto] {
	repo, err := newProblemSetsRepository(psm.kvs)
	if err != nil {
		return shpanstream.NewErrorStream[shpankids.ProblemSetDto](err)
	}
	// Find the problems in repo
	return shpanstream.MapStream(repo.Stream(ctx), mapProblemSetDbToDto)
}

func (psm *Manager) ListProblemsForProblemSet(ctx context.Context, problemSetId string, includingArchived bool) shpanstream.Stream[shpankids.ProblemDto] {
	// Get the user email from the context
	repo, err := newProblemSetProblemsRepository(ctx, psm.kvs, problemSetId)
	if err != nil {
		return shpanstream.NewErrorStream[shpankids.ProblemDto](err)
	}

	var s shpanstream.Stream[functional.Entry[string, dbProblem]]
	// Find the problems in repo
	if includingArchived {
		s = repo.StreamIncludingArchived(ctx)

	} else {
		s = repo.Stream(ctx)
	}
	return shpanstream.MapStream(s, mapProblemDbToDto)
}

func (psm *Manager) GenerateNewProblems(
	ctx context.Context,
	problemSetId string,
	additionalRequestText string,
) shpanstream.Stream[openapi.ApiProblemForEdit] {
	ps, err := psm.getProblemSet(ctx, problemSetId)
	if err != nil {
		return shpanstream.NewErrorStream[openapi.ApiProblemForEdit](err)
	}

	return ai.GenerateProblems(
		ctx,
		*ps,
		shpanstream.MapStream(
			psm.ListProblemsForProblemSet(ctx, problemSetId, true),
			api.ToApiProblemForEdit,
		),
		additionalRequestText,
	)
}

func (psm *Manager) RefineProblems(
	ctx context.Context,
	problemSetId string,
	origProblems shpanstream.Stream[openapi.ApiProblemForEdit],
	refineInstructions string,
) shpanstream.Stream[openapi.ApiProblemForEdit] {
	ps, err := psm.getProblemSet(ctx, problemSetId)
	if err != nil {
		return shpanstream.NewErrorStream[openapi.ApiProblemForEdit](err)
	}
	return ai.RefineProblems(
		ctx,
		*ps,
		origProblems,
		refineInstructions,
	)

}

func (psm *Manager) SubmitProblemAnswer(
	ctx context.Context,
	problemSetId string,
	problemId string,
	forDate datekvs.Date,
	answerId string,
) (bool, string, *shpankids.ProblemDto, error) {
	pRepo, err := newProblemSetProblemsRepository(ctx, psm.kvs, problemSetId)
	if err != nil {
		return false, "", nil, err
	}
	// Find the problems in repo
	dbP, err := pRepo.Get(ctx, problemId)
	if err != nil {
		return false, "", nil, err
	}

	correctAnswerId := functional.FindKeyInMap(dbP.Answers, func(a *dbProblemAnswer) bool {
		return a.Correct
	})
	if correctAnswerId == nil {
		return false, "", nil, fmt.Errorf("no correct answer found for problem %s", problemId)
	}

	solRepo, err := newProblemsSolutionsRepository(ctx, psm.kvs, problemSetId)
	if err != nil {
		return false, "", nil, err
	}
	dbPs := dbProblemSolution{
		SelectedAnswerId: answerId,
		Correct:          answerId == *correctAnswerId,
	}
	err = solRepo.Set(ctx, forDate, problemId, dbPs)
	if err != nil {
		return false, "", nil, err
	}
	// todo:amit:tx?
	err = pRepo.Archive(ctx, problemId)
	if err != nil {
		return false, "", nil, err
	}

	return dbPs.Correct, *correctAnswerId, mapProblemDbToDto(
		&functional.Entry[string, dbProblem]{Key: problemId, Value: dbP},
	), nil

}

func (psm *Manager) ListProblemSetSolutionsForDate(
	ctx context.Context,
	problemSetId string,
	forDate datekvs.Date,
) shpanstream.Stream[shpankids.ProblemSolutionDto] {
	sr, err := newProblemsSolutionsRepository(ctx, psm.kvs, problemSetId)
	if err != nil {
		return shpanstream.NewErrorStream[shpankids.ProblemSolutionDto](err)
	}
	return shpanstream.MapStream(
		sr.StreamAllForDate(ctx, forDate),
		func(e *functional.Entry[string, dbProblemSolution]) *shpankids.ProblemSolutionDto {
			return &shpankids.ProblemSolutionDto{
				ProblemId:        e.Key,
				SelectedAnswerId: e.Value.SelectedAnswerId,
				Correct:          e.Value.Correct,
			}
		},
	)
}

func (psm *Manager) ListProblemsSolutions(
	ctx context.Context,
	problemSetId string,
) shpanstream.Stream[openapi.ApiUserProblemSolution] {
	type titleAndCorrectAnswerId struct {
		Title           string
		CorrectAnswerId string
	}

	sr, err := newProblemsSolutionsRepository(ctx, psm.kvs, problemSetId)
	if err != nil {
		return shpanstream.NewErrorStream[openapi.ApiUserProblemSolution](err)
	}

	// todo:amit:bad, all in memory :(
	allProblems, err := psm.ListProblemsForProblemSet(
		ctx,
		problemSetId,
		true,
	).CollectFilterNil(ctx)
	if err != nil {
		return shpanstream.NewErrorStream[openapi.ApiUserProblemSolution](err)
	}
	problemMap := functional.SliceToMapKeyAndValueNoErr(
		allProblems,
		func(p shpankids.ProblemDto) string {
			return p.ProblemId
		}, func(p shpankids.ProblemDto) titleAndCorrectAnswerId {
			first := functional.FindFirst(p.Answers, func(a shpankids.ProblemAnswerDto) bool {
				return a.Correct
			})
			if first == nil {
				return titleAndCorrectAnswerId{
					Title:           p.Title,
					CorrectAnswerId: "",
				}
			}
			return titleAndCorrectAnswerId{
				Title:           p.Title,
				CorrectAnswerId: first.Id,
			}
		})

	return shpanstream.MapStream(
		sr.Stream(ctx),
		func(
			e *datekvs.DatedRecord[functional.Entry[string, dbProblemSolution]],
		) *openapi.ApiUserProblemSolution {
			return &openapi.ApiUserProblemSolution{
				ProblemId:            e.Value.Key,
				CorrectAnswerId:      problemMap[e.Value.Key].CorrectAnswerId,
				ProblemTitle:         problemMap[e.Value.Key].Title,
				SolvedDate:           e.Date.Time,
				UserProvidedAnswerId: e.Value.Value.SelectedAnswerId,
				Correct:              e.Value.Value.Correct,
			}
		},
	)
}

func (psm *Manager) GetProblem(
	ctx context.Context,
	problemSetId string,
	problemId string,
) (*shpankids.ProblemDto, error) {
	pRepo, err := newProblemSetProblemsRepository(ctx, psm.kvs, problemSetId)
	if err != nil {
		return nil, err
	}
	// Find the problems in repo
	dbP, err := pRepo.GetIncludingArchived(ctx, problemId)
	if err != nil {
		return nil, err
	}
	return mapProblemDbToDto(&functional.Entry[string, dbProblem]{Key: problemId, Value: dbP}), nil
}

func mapProblemSetDbToDto(e *functional.Entry[string, dbProblemSet]) *shpankids.ProblemSetDto {
	return &shpankids.ProblemSetDto{
		ProblemSetId: e.Key,
		Title:        e.Value.Title,
		Description:  e.Value.Description,
		Created:      e.Value.Created,
	}
}

func mapProblemDbToDto(e *functional.Entry[string, dbProblem]) *shpankids.ProblemDto {
	return &shpankids.ProblemDto{
		ProblemId:   e.Key,
		Title:       e.Value.Title,
		Description: e.Value.Description,
		Created:     e.Value.Created,
		Hints:       e.Value.Hints,
		Explanation: e.Value.Explanation,
		Answers:     functional.MapToSliceNoErr(e.Value.Answers, mapProblemAnswerDbToDto),
	}
}

func mapProblemAnswerDbToDto(problemId string, a dbProblemAnswer) shpankids.ProblemAnswerDto {
	return shpankids.ProblemAnswerDto{
		Id:          problemId,
		Title:       a.Title,
		Description: a.Description,
		Correct:     a.Correct,
	}
}

func (psm *Manager) getProblemSet(
	ctx context.Context,
	problemSetId string,
) (*shpankids.ProblemSetDto, error) {
	psRepo, err := newProblemSetsRepository(psm.kvs)
	if err != nil {
		return nil, err
	}
	dbPs, err := psRepo.Get(ctx, problemSetId)
	if err != nil {
		return nil, err
	}
	return mapProblemSetDbToDto(&functional.Entry[string, dbProblemSet]{Key: problemSetId, Value: dbPs}), nil

}
