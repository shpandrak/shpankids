package api

import (
	"context"
	"shpankids/domain/task"
	"shpankids/infra/util/functional"
	"shpankids/openapi"
	"shpankids/shpankids"
	"time"
)

type OapiServerApiImpl struct {
	userSessionManager shpankids.UserSessionManager
	userManager        shpankids.UserManager
	taskManager        task.Manager
}

func NewOapiServerApiImpl(
	userSessionManager shpankids.UserSessionManager,
	userManager shpankids.UserManager,
	taskManager task.Manager,

) *OapiServerApiImpl {
	return &OapiServerApiImpl{
		userSessionManager: userSessionManager,
		userManager:        userManager,
		taskManager:        taskManager,
	}
}

func (oa *OapiServerApiImpl) ListTasks(
	ctx context.Context,
	_ openapi.ListTasksRequestObject,
) (openapi.ListTasksResponseObject, error) {
	forDate := GetTodayForDate()
	taskList, err := oa.taskManager.GetTasksForDate(ctx, forDate)
	if err != nil {
		return nil, err
	}
	return openapi.ListTasks200JSONResponse(
		functional.MapSliceNoErr(taskList, func(t task.Task) openapi.ApiTask {
			return openapi.ApiTask{
				Id:          t.Id,
				Title:       t.Title,
				Description: t.Description,
				ForDate:     forDate,
				Status:      openapi.ApiTaskStatus(t.Status),
			}
		})), nil

}

func GetTodayForDate() time.Time {
	return time.Now().Truncate(24 * time.Hour)
}
