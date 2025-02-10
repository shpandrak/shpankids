package task

import (
	"cloud.google.com/go/firestore"
	"context"
	"errors"
	"google.golang.org/api/iterator"
	"shpankids/infra/util/functional"
	"shpankids/shpankids"
	"time"
)

type managerImpl struct {
	fs                 *firestore.Client
	userSessionManager shpankids.UserSessionManager
	sessionManager     shpankids.SessionManager
	familyManager      shpankids.FamilyManager
}

func NewTaskManager(
	fs *firestore.Client,
	userSessionManager shpankids.UserSessionManager,
	familyManager shpankids.FamilyManager,
	sessionManager shpankids.SessionManager,
) shpankids.Manager {
	return &managerImpl{
		fs:                 fs,
		userSessionManager: userSessionManager,
		familyManager:      familyManager,
		sessionManager:     sessionManager,
	}
}

func (m *managerImpl) GetTasksForDate(ctx context.Context, forDate time.Time) ([]shpankids.Task, error) {
	userId, err := m.userSessionManager(ctx)
	if err != nil {
		return nil, err
	}

	s, err := m.sessionManager.Get(ctx, *userId)
	if err != nil {
		return nil, err
	}
	familyTasks, err := m.familyManager.ListFamilyTasks(ctx, s.FamilyId)
	if err != nil {
		return nil, err
	}

	tasks := functional.MapSliceNoErr(familyTasks, func(ft shpankids.FamilyTaskDto) *shpankids.Task {
		return &shpankids.Task{
			Id:          ft.TaskId,
			Title:       ft.Title,
			Description: ft.Description,
			Status:      shpankids.StatusOpen,
		}
	})

	tasksById := functional.SliceToMapNoErr(tasks, func(t *shpankids.Task) string {
		return t.Id
	})
	docIter := m.fs.
		Collection("users").
		Doc(*userId).
		Collection("tasks-" + forDate.Format(time.DateOnly)).
		Documents(ctx)

	for {
		doc, err := docIter.Next()
		if err != nil {
			if !errors.Is(err, iterator.Done) {
				return nil, err
			}
			break
		}
		foundTask, found := tasksById[doc.Ref.ID]
		if found {
			stsStr, ok := doc.Data()["status"]
			if ok {
				foundTask.Status = shpankids.Status(stsStr.(string))
			}
		}
	}

	return functional.MapSliceUnPtr(functional.MapValues(tasksById)), nil
}

func (m *managerImpl) UpdateTaskStatus(ctx context.Context, forDay time.Time, taskId string, status shpankids.Status, comment string) error {
	userId, err := m.userSessionManager(ctx)
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"status": status,
	}
	if comment != "" {
		data["comment"] = comment
	}
	_, err = m.fs.
		Collection("users").
		Doc(*userId).
		Collection("tasks-"+forDay.Format(time.DateOnly)).
		Doc(taskId).
		Set(ctx, data)
	return err

}
