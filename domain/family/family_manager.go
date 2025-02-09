package family

import (
	"context"
	"shpankids/infra/database/kvstore"
	"shpankids/infra/util/functional"
	"shpankids/shpankids"
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
	// Get the user email from the context
	_, err := m.userSessionManager(ctx)
	if err != nil {
		return nil, err
	}
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

func mapFamilyDto(fam *dbFamily) *shpankids.FamilyDto {
	return &shpankids.FamilyDto{
		Id:   fam.Id,
		Name: fam.Name,
	}
}
func (m *Manager) CreateFamilyTask(ctx context.Context, familyId string, familyTask shpankids.FamilyTaskDto) error {
	// Get the user email from the context
	_, err := m.userSessionManager(ctx)
	if err != nil {
		return err
	}
	familyTasksStore, err := m.kvs.CreateSpaceStore(ctx, []string{"families", familyId})
	if err != nil {
		return err
	}
	repo := newFamilyTaskRepository(familyTasksStore)
	// Create the family task in repo
	return repo.Set(ctx, familyTask.TaskId, dbFamilyTask{
		Title:       familyTask.Title,
		Description: familyTask.Description,
		MemberIds:   familyTask.MemberIds,
	})
}
func (m *Manager) FamilyTasks(ctx context.Context, familyId string) ([]shpankids.FamilyTaskDto, error) {
	// Get the user email from the context
	_, err := m.userSessionManager(ctx)
	if err != nil {
		return nil, err
	}
	familyTasksStore, err := m.kvs.CreateSpaceStore(ctx, []string{"families", familyId})
	if err != nil {
		return nil, err
	}
	repo := newFamilyTaskRepository(familyTasksStore)
	// Find the family tasks in repo
	dbTasks, err := repo.List(ctx)
	if err != nil {
		return nil, err
	}
	return functional.MapToSliceNoErr(dbTasks, func(taskId string, dbFt dbFamilyTask) shpankids.FamilyTaskDto {
		return shpankids.FamilyTaskDto{
			TaskId:      taskId,
			Title:       dbFt.Title,
			Description: dbFt.Description,
			MemberIds:   dbFt.MemberIds,
		}
	}), nil
}

func (m *Manager) CreateFamily(
	ctx context.Context,
	familyId string,
	familyName string,
	memberUserIds []string,
) error {
	// Get the user email from the context
	loggedInUserEmail, err := m.userSessionManager(ctx)
	if err != nil {
		return err
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
			Members: append([]dbFamilyMember{
				{
					UserId: *loggedInUserEmail,
					Role:   RoleAdmin,
				},
			}, functional.MapSliceNoErr(memberUserIds, func(userId string) dbFamilyMember {
				return dbFamilyMember{
					UserId: userId,
					Role:   RoleMember,
				}
			})...),
		},
	)
}
