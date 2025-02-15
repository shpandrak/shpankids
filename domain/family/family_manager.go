package family

import (
	"context"
	"fmt"
	"shpankids/infra/database/kvstore"
	"shpankids/infra/shpanstream"
	"shpankids/infra/util/functional"
	"shpankids/internal/infra/util"
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
		Status:      shpankids.FamilyTaskStatusActive,
		Created:     familyTask.Created,
		StatusDate:  familyTask.Created,
	})
}

func (m *Manager) CreateFamilyProblem(
	ctx context.Context,
	familyId string,
	forUserId string,
	familyProblem shpankids.FamilyProblemDto,
) error {
	if familyProblem.ProblemId == "" {
		return fmt.Errorf("task id is required")
	}
	if familyProblem.Title == "" {
		return util.BadInputError(fmt.Errorf("title is required"))
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

	if _, ok := famMembersSet[forUserId]; !ok {
		return util.BadInputError(fmt.Errorf("user %s is not part of the family %s", forUserId, f.Name))
	}

	repo, err := newFamilyProblemsRepository(ctx, m.kvs, familyId, forUserId)
	if err != nil {
		return err
	}

	// Create the family task in repo
	familyProblem.Created = time.Now()
	return repo.Set(ctx, familyProblem.ProblemId, dbFamilyProblem{
		Title:       familyProblem.Title,
		Description: familyProblem.Description,
		Created:     familyProblem.Created,
		Hints:       familyProblem.Hints,
		Explanation: familyProblem.Explanation,
		Alternatives: functional.MapSliceNoErr(familyProblem.Alternatives, func(a shpankids.ProblemAlternativeDto) dbProblemAlternative {
			return dbProblemAlternative{
				Title:       a.Title,
				Description: a.Description,
				Correct:     a.Correct,
			}
		}),
		Status:     shpankids.FamilyTaskStatusActive,
		StatusDate: familyProblem.Created,
	})
}

func (m *Manager) ListFamilyTasks(ctx context.Context, familyId string) shpanstream.Stream[shpankids.FamilyTaskDto] {
	// Get the user email from the context
	_, err := m.userSessionManager(ctx)
	if err != nil {
		return shpanstream.NewErrorStream[shpankids.FamilyTaskDto](err)
	}
	repo, err := newFamilyTaskRepository(ctx, m.kvs, familyId)
	if err != nil {
		return shpanstream.NewErrorStream[shpankids.FamilyTaskDto](err)
	}
	// Find the family tasks in repo
	return shpanstream.MapStream(
		repo.Stream(ctx),
		func(e *functional.Entry[string, dbFamilyTask]) *shpankids.FamilyTaskDto {
			return &shpankids.FamilyTaskDto{
				TaskId:      e.Key,
				Title:       e.Value.Title,
				Description: e.Value.Description,
				MemberIds:   e.Value.MemberIds,
				Status:      e.Value.Status,
				StatusDate:  e.Value.StatusDate,
				Created:     e.Value.Created,
			}
		},
	)
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
	ft.Status = shpankids.FamilyTaskStatusDeleted
	ft.StatusDate = time.Now()
	return repo.Set(ctx, familyTaskId, ft)
}
