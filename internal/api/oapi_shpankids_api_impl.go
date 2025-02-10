package api

import (
	"context"
	"shpankids/infra/util/functional"
	"shpankids/openapi"
	"shpankids/shpankids"
	"time"
)

type OapiServerApiImpl struct {
	userSessionManager shpankids.UserSessionManager
	userManager        shpankids.UserManager
	taskManager        shpankids.Manager
	familyManager      shpankids.FamilyManager
	sessionManager     shpankids.SessionManager
}

func NewOapiServerApiImpl(
	userSessionManager shpankids.UserSessionManager,
	userManager shpankids.UserManager,
	taskManager shpankids.Manager,
	familyManager shpankids.FamilyManager,
	sessionManager shpankids.SessionManager,

) *OapiServerApiImpl {
	return &OapiServerApiImpl{
		userSessionManager: userSessionManager,
		userManager:        userManager,
		taskManager:        taskManager,
		familyManager:      familyManager,
		sessionManager:     sessionManager,
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
		functional.MapSliceNoErr(taskList, func(t shpankids.Task) openapi.ApiTask {
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
