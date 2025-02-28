package family

import (
	"context"
	"fmt"
	"shpankids/domain/problemset"
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
	sessionManager     shpankids.SessionManager
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

func (m *Manager) GetProblemSetManagerForUser(ctx context.Context, forUserId string) (shpankids.ProblemSetManager, error) {
	loggedInUserId, err := m.userSessionManager(ctx)
	if err != nil {
		return nil, err
	}

	// todo:amit: load session once per request, everyone needs it...
	s, err := m.sessionManager.Get(ctx, *loggedInUserId)
	if err != nil {
		return nil, err
	}

	// todo:amit: load family once for multiple, this is crazy, store in ctx?
	f, err := m.GetFamily(ctx, s.FamilyId)
	if err != nil {
		return nil, err
	}

	// Check if the user is and admin of the family
	isAllowed := slices.ContainsFunc(f.Members, func(member shpankids.FamilyMemberDto) bool {
		return member.UserId == *loggedInUserId && member.Role == shpankids.RoleAdmin
	})
	if !isAllowed {
		return nil, util.ForbiddenError(fmt.Errorf("not an admin for family %s", f.Name))
	}

	famUserKvs, err := createFamilyUserRootStore(ctx, m.kvs, s.FamilyId, forUserId)
	if err != nil {
		return nil, err
	}
	return problemset.NewProblemSetManager(famUserKvs), nil
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
