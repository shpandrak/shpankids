package family

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"shpankids/infra/database/datekvs"
	"shpankids/infra/database/kvstore"
	"shpankids/infra/shpanstream"
	"shpankids/infra/util/functional"
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
		dbProblemSet{
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

	correctAnswerId := functional.FindKeyInMap(dbP.Answers, func(a *dbProblemAnswer) bool {
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
		&functional.Entry[string, dbProblem]{Key: problemId, Value: dbP},
	), nil

}

func (m *Manager) getProblem(
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
	dbP, err := pRepo.Get(ctx, problemId)
	if err != nil {
		return nil, err
	}
	return mapFamilyProblemDbToDto(&functional.Entry[string, dbProblem]{Key: problemId, Value: dbP}), nil
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
) shpanstream.Stream[shpankids.FamilyProblemDto] {
	// Get the user email from the context
	repo, err := newFamilyProblemsRepository(ctx, m.kvs, familyId, userId, problemSetId)
	if err != nil {
		return shpanstream.NewErrorStream[shpankids.FamilyProblemDto](err)
	}
	// Find the problems in repo
	return shpanstream.MapStream(repo.StreamIncludingArchived(ctx), mapFamilyProblemDbToDto)

}

func mapFamilyProblemDbToDto(e *functional.Entry[string, dbProblem]) *shpankids.FamilyProblemDto {
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

func mapFamilyProblemSetDbToDto(e *functional.Entry[string, dbProblemSet]) *shpankids.FamilyProblemSetDto {
	return &shpankids.FamilyProblemSetDto{
		ProblemSetId: e.Key,
		Title:        e.Value.Title,
		Description:  e.Value.Description,
		Created:      e.Value.Created,
		Status:       e.Value.Status,
		StatusDate:   e.Value.StatusDate,
	}
}

func (m *Manager) GenerateNewProblems(
	ctx context.Context,
	familyId string,
	userId string,
	problemSetId string,
	additionalRequestText string,
) shpanstream.Stream[openapi.ApiProblemForEdit] {
	psRepo, err := newProblemSetsRepository(ctx, m.kvs, familyId, userId)
	if err != nil {
		return shpanstream.NewErrorStream[openapi.ApiProblemForEdit](err)
	}
	dbPs, err := psRepo.Get(ctx, problemSetId)
	if err != nil {
		return shpanstream.NewErrorStream[openapi.ApiProblemForEdit](err)
	}
	return generateProblems(
		ctx,
		userId,
		*mapFamilyProblemSetDbToDto(&functional.Entry[string, dbProblemSet]{Key: problemSetId, Value: dbPs}),
		m.ListProblemsForProblemSet(ctx, familyId, userId, problemSetId),
		additionalRequestText,
	)
}
func mapFamilyProblemAlternativeDbToDto(problemId string, a dbProblemAnswer) shpankids.ProblemAnswerDto {
	return shpankids.ProblemAnswerDto{
		Id:          problemId,
		Title:       a.Title,
		Description: a.Description,
		Correct:     a.Correct,
	}
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
	return shpanstream.MapStream(sr.GetAllForDate(ctx, forDate), func(e *functional.Entry[string, dbProblemSolution]) *shpankids.ProblemSolutionDto {
		return &shpankids.ProblemSolutionDto{
			ProblemId:        e.Key,
			SelectedAnswerId: e.Value.SelectedAnswerId,
			Correct:          e.Value.Correct,
		}
	})
}
