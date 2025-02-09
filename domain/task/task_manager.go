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
}

func NewTaskManager(fs *firestore.Client, userSessionManager shpankids.UserSessionManager) Manager {
	return &managerImpl{
		fs:                 fs,
		userSessionManager: userSessionManager,
	}
}

func (m *managerImpl) GetTasksForDate(ctx context.Context, forDate time.Time) ([]Task, error) {
	userId, err := m.userSessionManager(ctx)
	if err != nil {
		return nil, err
	}

	tasks := []*Task{
		{
			Id:          "1",
			Title:       "Do homework",
			Description: "Check the app to see what homework needs to be done",
			Status:      StatusOpen,
		},
		{
			Id:     "2",
			Title:  "Put lunchbox in the dishwasher",
			Status: StatusOpen,
		},
	}

	tasksById := functional.SliceToMapNoErr(tasks, func(t *Task) string {
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
				foundTask.Status = Status(stsStr.(string))
			}
		}
	}

	return functional.MapSliceUnPtr(functional.MapValues(tasksById)), nil
}

func (m *managerImpl) UpdateTaskStatus(ctx context.Context, forDay time.Time, taskId string, status Status, comment string) error {
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
