package family

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"log/slog"
	"shpankids/domain/ai"
	"shpankids/domain/problemset"
	"shpankids/infra/database/datekvs"
	"shpankids/infra/database/kvstore"
	"shpankids/infra/shpanstream"
	"shpankids/infra/util/functional"
	"shpankids/internal/api"
	"shpankids/internal/infra/util"
	"shpankids/openapi"
	"shpankids/shpankids"
	"slices"
	"time"
)

type Manager struct {
	familyRepository   repository
	userSessionManager shpankids.UserSessionManager
	kvs                kvstore.RawJsonStore
}

func NewFamilyManager(
	kvs kvstore.RawJsonStore,
	userSessionManager shpankids.UserSessionManager,
) *Manager {
	return &Manager{
		familyRepository:   newFamilyRepository(kvs),
		userSessionManager: userSessionManager,
		kvs:                kvs,
	}
}

func (m *Manager) FindFamily(ctx context.Context, familyId string) (*shpankids.FamilyDto, error) {
	// Find the family in repo
	dbFam, err := m.familyRepository.Find(ctx, familyId)
	if err != nil {
		return nil, err
	}
	if dbFam == nil {
		return nil, nil
	}
	return mapFamilyDto(dbFam), nil

}

func (m *Manager) GetFamily(ctx context.Context, familyId string) (*shpankids.FamilyDto, error) {
	// Find the family in repo
	dbFam, err := m.familyRepository.Get(ctx, familyId)
	if err != nil {
		return nil, err
	}
	return mapFamilyDto(&dbFam), nil

}

func mapFamilyDto(fam *dbFamily) *shpankids.FamilyDto {
	return &shpankids.FamilyDto{
		Id:         fam.Id,
		Name:       fam.Name,
		OwnerEmail: fam.CreatedBy,
		CreatedOn:  fam.CreatedAt,
		Members: functional.MapSliceNoErr(fam.Members, func(member dbFamilyMember) shpankids.FamilyMemberDto {
			return shpankids.FamilyMemberDto{
				UserId: member.UserId,
				Role:   member.Role,
			}
		}),
	}
}

func (m *Manager) CreateFamilyTask(ctx context.Context, familyId string, familyTask shpankids.FamilyTaskDto) error {
	if familyTask.TaskId == "" {
		return fmt.Errorf("task id is required")
	}
	if familyTask.Title == "" {
		return util.BadInputError(fmt.Errorf("title is required"))
	}
	if len(familyTask.MemberIds) == 0 {
		return util.BadInputError(fmt.Errorf("at least one member is required for a task"))
	}

	// Get the user email from the context
	uId, err := m.userSessionManager(ctx)
	if err != nil {
		return err
	}

	f, err := m.GetFamily(ctx, familyId)
	if err != nil {
		return err
	}

	// Check if the user is and admin of the family
	isAdmin := slices.ContainsFunc(f.Members, func(member shpankids.FamilyMemberDto) bool {
		return member.UserId == *uId && member.Role == shpankids.RoleAdmin
	})
	if !isAdmin {
		return util.ForbiddenError(fmt.Errorf("only admin can create family tasks for family %s", f.Name))
	}

	// Check if all task members are part of the family
	famMembersSet := functional.SliceToSetExtractKeyNoErr(f.Members, func(member shpankids.FamilyMemberDto) string {
		return member.UserId
	})

	for _, memberId := range familyTask.MemberIds {
		if _, ok := famMembersSet[memberId]; !ok {
			return util.BadInputError(fmt.Errorf("member %s is not part of the family %s", memberId, f.Name))
		}
	}

	repo, err := newFamilyTaskRepository(ctx, m.kvs, familyId)
	if err != nil {
		return err
	}
	// Create the family task in repo

	familyTask.Created = time.Now()
	return repo.Set(ctx, familyTask.TaskId, dbFamilyTask{
		Title:       familyTask.Title,
		Description: familyTask.Description,
		MemberIds:   familyTask.MemberIds,
		Status:      shpankids.FamilyAssignmentStatusActive,
		Created:     familyTask.Created,
		StatusDate:  familyTask.Created,
	})
}
func (m *Manager) CreateProblemSet(ctx context.Context, familyId string, forUserId string, familyProblemSet shpankids.CreateProblemSetDto) error {
	psRepo, err := newProblemSetsRepository(ctx, m.kvs, familyId, forUserId)
	if err != nil {
		return err
	}
	createTime := time.Now()
	// Create the family task in repo
	return psRepo.Set(
		ctx,
		familyProblemSet.ProblemSetId,
		problemset.DbProblemSet{
			Title:       familyProblemSet.Title,
			Description: familyProblemSet.Description,
			Created:     createTime,
			Status:      shpankids.FamilyAssignmentStatusActive,
			StatusDate:  createTime,
		})
}

func (m *Manager) CreateProblemsInSet(
	ctx context.Context,
	familyId string,
	forUserId string,
	problemSetId string,
	familyProblem []shpankids.CreateProblemDto,
) error {
	// Get the user email from the context
	uId, err := m.userSessionManager(ctx)
	if err != nil {
		return err
	}

	f, err := m.GetFamily(ctx, familyId)
	if err != nil {
		return err
	}

	// Check if the user is and admin of the family
	isAdmin := slices.ContainsFunc(f.Members, func(member shpankids.FamilyMemberDto) bool {
		return member.UserId == *uId && member.Role == shpankids.RoleAdmin
	})
	if !isAdmin {
		return util.ForbiddenError(fmt.Errorf("only admin can create family tasks for family %s", f.Name))
	}
	// Check if all task members are part of the family
	famMembersSet := functional.SliceToSetExtractKeyNoErr(f.Members, func(member shpankids.FamilyMemberDto) string {
		return member.UserId
	})

	if _, ok := famMembersSet[forUserId]; !ok {
		return util.BadInputError(fmt.Errorf("user %s is not part of the family %s", forUserId, f.Name))
	}

	repo, err := newFamilyProblemsRepository(ctx, m.kvs, familyId, forUserId, problemSetId)
	if err != nil {
		return err
	}
	createdTime := time.Now()
	for _, p := range familyProblem {
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

		dbAnswers := make(map[string]problemset.DbProblemAnswer, len(p.Answers))
		for idx, a := range p.Answers {
			dbAnswers[fmt.Sprintf("%d", idx)] = problemset.DbProblemAnswer{
				Title:       a.Title,
				Description: a.Description,
				Correct:     a.Correct,
			}
		}

		// Create the family task in repo
		err = repo.Set(ctx, uuid.NewString(), problemset.DbProblem{
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

func (m *Manager) SubmitProblemAnswer(
	ctx context.Context,
	familyId string,
	userId string,
	problemSetId string,
	problemId string,
	forDate datekvs.Date,
	answerId string,
) (bool, string, *shpankids.FamilyProblemDto, error) {

	pRepo, err := newFamilyProblemsRepository(ctx, m.kvs, familyId, userId, problemSetId)
	if err != nil {
		return false, "", nil, err
	}
	// Find the problems in repo
	dbP, err := pRepo.Get(ctx, problemId)
	if err != nil {
		return false, "", nil, err
	}

	correctAnswerId := functional.FindKeyInMap(dbP.Answers, func(a *problemset.DbProblemAnswer) bool {
		return a.Correct
	})
	if correctAnswerId == nil {
		return false, "", nil, fmt.Errorf("no correct answer found for problem %s", problemId)
	}

	solRepo, err := newFamilyProblemsSolutionsRepository(ctx, m.kvs, familyId, userId, problemSetId)
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

	return dbPs.Correct, *correctAnswerId, mapFamilyProblemDbToDto(
		&functional.Entry[string, problemset.DbProblem]{Key: problemId, Value: dbP},
	), nil

}

func (m *Manager) GetProblem(
	ctx context.Context,
	familyId string,
	userId string,
	problemSetId string,
	problemId string,
) (*shpankids.FamilyProblemDto, error) {
	pRepo, err := newFamilyProblemsRepository(ctx, m.kvs, familyId, userId, problemSetId)
	if err != nil {
		return nil, err
	}
	// Find the problems in repo
	dbP, err := pRepo.GetIncludingArchived(ctx, problemId)
	if err != nil {
		return nil, err
	}
	return mapFamilyProblemDbToDto(&functional.Entry[string, problemset.DbProblem]{Key: problemId, Value: dbP}), nil
}

func (m *Manager) ListFamilyTasks(ctx context.Context, familyId string) shpanstream.Stream[shpankids.FamilyTaskDto] {
	repo, err := newFamilyTaskRepository(ctx, m.kvs, familyId)
	if err != nil {
		return shpanstream.NewErrorStream[shpankids.FamilyTaskDto](err)
	}
	// Find the family tasks in repo
	return shpanstream.MapStream(repo.Stream(ctx), mapFamilyTaskDbToDto)
}

func mapFamilyTaskDbToDto(e *functional.Entry[string, dbFamilyTask]) *shpankids.FamilyTaskDto {
	return &shpankids.FamilyTaskDto{
		TaskId:      e.Key,
		Title:       e.Value.Title,
		Description: e.Value.Description,
		MemberIds:   e.Value.MemberIds,
		Status:      e.Value.Status,
		StatusDate:  e.Value.StatusDate,
		Created:     e.Value.Created,
	}
}

func (m *Manager) CreateFamily(
	ctx context.Context,
	familyId string,
	familyName string,
	memberUserIds []string,
	adminUserIds []string,
) error {
	// Get the user email from the context
	loggedInUserEmail, err := m.userSessionManager(ctx)
	if err != nil {
		return err
	}

	famMembersByUserId := map[string]dbFamilyMember{}
	famMembersByUserId[*loggedInUserEmail] = dbFamilyMember{
		UserId: *loggedInUserEmail,
		Role:   shpankids.RoleAdmin,
	}
	for _, memberId := range memberUserIds {
		famMembersByUserId[memberId] = dbFamilyMember{
			UserId: memberId,
			Role:   shpankids.RoleMember,
		}
	}
	for _, adminId := range adminUserIds {
		famMembersByUserId[adminId] = dbFamilyMember{
			UserId: adminId,
			Role:   shpankids.RoleAdmin,
		}
	}

	// Create the family in repo
	return m.familyRepository.Set(
		ctx,
		familyId,
		dbFamily{
			Id:        familyId,
			Name:      familyName,
			CreatedBy: *loggedInUserEmail,
			CreatedAt: time.Now(),
			Members:   functional.MapValues(famMembersByUserId),
		},
	)
}

func (m *Manager) DeleteFamilyTask(ctx context.Context, familyId string, familyTaskId string) error {
	// Get the user email from the context
	uId, err := m.userSessionManager(ctx)
	if err != nil {
		return err
	}

	f, err := m.GetFamily(ctx, familyId)
	if err != nil {
		return err
	}

	// Check if the user is and admin of the family
	isAdmin := slices.ContainsFunc(f.Members, func(member shpankids.FamilyMemberDto) bool {
		return member.UserId == *uId && member.Role == shpankids.RoleAdmin
	})
	if !isAdmin {
		return util.ForbiddenError(fmt.Errorf("only admin can delete family tasks for family %s", f.Name))
	}
	repo, err := newFamilyTaskRepository(ctx, m.kvs, familyId)
	if err != nil {
		return err
	}
	ft, err := repo.Get(ctx, familyTaskId)
	if err != nil {
		return err
	}
	ft.Status = shpankids.FamilyAssignmentStatusDeleted
	ft.StatusDate = time.Now()
	return repo.Set(ctx, familyTaskId, ft)
}

func (m *Manager) ListProblemSetsForUser(
	ctx context.Context,
	familyId string,
	userId string,
) shpanstream.Stream[shpankids.FamilyProblemSetDto] {

	repo, err := newProblemSetsRepository(ctx, m.kvs, familyId, userId)
	if err != nil {
		return shpanstream.NewErrorStream[shpankids.FamilyProblemSetDto](err)
	}
	// Find the problems in repo
	return shpanstream.MapStream(repo.Stream(ctx), mapFamilyProblemSetDbToDto)
}

func (m *Manager) ListProblemsForProblemSet(
	ctx context.Context,
	familyId string,
	userId string,
	problemSetId string,
	includeArchived bool,
) shpanstream.Stream[shpankids.FamilyProblemDto] {
	// Get the user email from the context
	repo, err := newFamilyProblemsRepository(ctx, m.kvs, familyId, userId, problemSetId)
	if err != nil {
		return shpanstream.NewErrorStream[shpankids.FamilyProblemDto](err)
	}

	var s shpanstream.Stream[functional.Entry[string, problemset.DbProblem]]
	// Find the problems in repo
	if includeArchived {
		s = repo.StreamIncludingArchived(ctx)

	} else {
		s = repo.Stream(ctx)
	}
	return shpanstream.MapStream(s, mapFamilyProblemDbToDto)

}

func mapFamilyProblemDbToDto(e *functional.Entry[string, problemset.DbProblem]) *shpankids.FamilyProblemDto {
	return &shpankids.FamilyProblemDto{
		ProblemId:   e.Key,
		Title:       e.Value.Title,
		Description: e.Value.Description,
		Created:     e.Value.Created,
		Hints:       e.Value.Hints,
		Explanation: e.Value.Explanation,
		Answers:     functional.MapToSliceNoErr(e.Value.Answers, mapFamilyProblemAlternativeDbToDto),
	}
}

func mapFamilyProblemSetDbToDto(e *functional.Entry[string, problemset.DbProblemSet]) *shpankids.FamilyProblemSetDto {
	return &shpankids.FamilyProblemSetDto{
		ProblemSetId: e.Key,
		Title:        e.Value.Title,
		Description:  e.Value.Description,
		Created:      e.Value.Created,
		Status:       e.Value.Status,
		StatusDate:   e.Value.StatusDate,
	}
}

func (m *Manager) getProblemSet(
	ctx context.Context,
	familyId string,
	userId string,
	problemSetId string,
) (*shpankids.FamilyProblemSetDto, error) {
	psRepo, err := newProblemSetsRepository(ctx, m.kvs, familyId, userId)
	if err != nil {
		return nil, err
	}
	dbPs, err := psRepo.Get(ctx, problemSetId)
	if err != nil {
		return nil, err
	}
	return mapFamilyProblemSetDbToDto(&functional.Entry[string, problemset.DbProblemSet]{Key: problemSetId, Value: dbPs}), nil

}

func (m *Manager) RefineProblems(
	ctx context.Context,
	familyId string,
	userId string,
	problemSetId string,
	origProblems shpanstream.Stream[openapi.ApiProblemForEdit],
	refineInstructions string,
) shpanstream.Stream[openapi.ApiProblemForEdit] {
	ps, err := m.getProblemSet(ctx, familyId, userId, problemSetId)
	if err != nil {
		return shpanstream.NewErrorStream[openapi.ApiProblemForEdit](err)
	}
	return ai.RefineProblems(
		ctx,
		userId,
		*ps,
		origProblems,
		refineInstructions,
	)

}

func (m *Manager) GenerateNewProblems(
	ctx context.Context,
	familyId string,
	userId string,
	problemSetId string,
	additionalRequestText string,
) shpanstream.Stream[openapi.ApiProblemForEdit] {
	ps, err := m.getProblemSet(ctx, familyId, userId, problemSetId)
	if err != nil {
		return shpanstream.NewErrorStream[openapi.ApiProblemForEdit](err)
	}

	return ai.GenerateProblems(
		ctx,
		userId,
		*ps,
		shpanstream.MapStream(
			m.ListProblemsForProblemSet(ctx, familyId, userId, problemSetId, true),
			api.ToApiProblemForEdit,
		),
		additionalRequestText,
	)
}
func mapFamilyProblemAlternativeDbToDto(problemId string, a problemset.DbProblemAnswer) shpankids.ProblemAnswerDto {
	return shpankids.ProblemAnswerDto{
		Id:          problemId,
		Title:       a.Title,
		Description: a.Description,
		Correct:     a.Correct,
	}
}

type titleAndCorrectAnswerId struct {
	Title           string
	CorrectAnswerId string
}

func (m *Manager) ListUserProblemsSolutions(
	ctx context.Context,
	familyId string,
	problemSetId string,
	userId string,
) shpanstream.Stream[openapi.ApiUserProblemSolution] {
	sr, err := newFamilyProblemsSolutionsRepository(ctx, m.kvs, familyId, userId, problemSetId)
	if err != nil {
		return shpanstream.NewErrorStream[openapi.ApiUserProblemSolution](err)
	}

	// todo:amit:bad, all in memory :(
	allProblems, err := m.ListProblemsForProblemSet(ctx, familyId, userId, problemSetId, true).CollectFilterNil(ctx)
	if err != nil {
		return shpanstream.NewErrorStream[openapi.ApiUserProblemSolution](err)
	}
	problemMap := functional.SliceToMapKeyAndValueNoErr(
		allProblems,
		func(p shpankids.FamilyProblemDto) string {
			return p.ProblemId
		}, func(p shpankids.FamilyProblemDto) titleAndCorrectAnswerId {
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

	filterNil, err := sr.Stream(ctx).CollectFilterNil(ctx)
	slog.Info(fmt.Sprintf("struff %v %v", filterNil, err))

	return shpanstream.MapStream(sr.Stream(ctx), func(e *datekvs.DatedRecord[functional.Entry[string, dbProblemSolution]]) *openapi.ApiUserProblemSolution {
		return &openapi.ApiUserProblemSolution{
			ProblemId:            e.Value.Key,
			CorrectAnswerId:      problemMap[e.Value.Key].CorrectAnswerId,
			ProblemTitle:         problemMap[e.Value.Key].Title,
			SolvedDate:           e.Date.Time,
			UserProvidedAnswerId: e.Value.Value.SelectedAnswerId,
			Correct:              e.Value.Value.Correct,
		}
	})

}

func (m *Manager) ListProblemSetSolutionsForDate(
	ctx context.Context,
	familyId string,
	userId string,
	problemSetId string,
	forDate datekvs.Date,
) shpanstream.Stream[shpankids.ProblemSolutionDto] {
	sr, err := newFamilyProblemsSolutionsRepository(ctx, m.kvs, familyId, userId, problemSetId)
	if err != nil {
		return shpanstream.NewErrorStream[shpankids.ProblemSolutionDto](err)
	}
	return shpanstream.MapStream(sr.StreamAllForDate(ctx, forDate), func(e *functional.Entry[string, dbProblemSolution]) *shpankids.ProblemSolutionDto {
		return &shpankids.ProblemSolutionDto{
			ProblemId:        e.Key,
			SelectedAnswerId: e.Value.SelectedAnswerId,
			Correct:          e.Value.Correct,
		}
	})
}
