package api

import (
	"context"
	"fmt"
	"shpankids/infra/util/functional"
	"shpankids/internal/infra/util"
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

func (oa *OapiServerApiImpl) GetStats(ctx context.Context, request openapi.GetStatsRequestObject) (openapi.GetStatsResponseObject, error) {
	from := time.Now().Truncate(24 * time.Hour)
	to := from

	if request.Params.From != nil {
		from = *request.Params.From
	}
	if request.Params.To != nil {
		to = *request.Params.To
	}
	if to.Before(from) {
		return nil, util.BadInputError(fmt.Errorf("to date is before from date"))
	}
	stats, err := oa.taskManager.GetTaskStats(ctx, from, to)
	if err != nil {
		return nil, err
	}
	return openapi.GetStats200JSONResponse(functional.MapSliceNoErr(
		stats,
		func(s shpankids.TaskStats) openapi.ApiTaskStats {
			return openapi.ApiTaskStats{
				UserId:          s.UserId,
				ForDate:         s.ForDate,
				TotalTasksCount: s.TotalTasksCount,
				DoneTasksCount:  s.DoneTasksCount,
			}
		},
	)), nil

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
