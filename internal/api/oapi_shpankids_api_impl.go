package api

import (
	"context"
	"fmt"
	"net/http"
	"shpankids/infra/database/datekvs"
	"shpankids/infra/shpanstream"
	"shpankids/infra/util/functional"
	"shpankids/internal/infra/util"
	"shpankids/openapi"
	"shpankids/shpankids"
	"time"
)

type OapiServerApiImpl struct {
	userSessionManager shpankids.UserSessionManager
	userManager        shpankids.UserManager
	taskManager        shpankids.TaskManager
	familyManager      shpankids.FamilyManager
	sessionManager     shpankids.SessionManager
}

func (oa *OapiServerApiImpl) GetStats(
	ctx context.Context,
	request openapi.GetStatsRequestObject,
) (openapi.GetStatsResponseObject, error) {
	from := *datekvs.NewDateFromTime(time.Now())
	to := from

	if request.Params.From != nil {
		from = *datekvs.NewDateFromTime(*request.Params.From)
	}
	if request.Params.To != nil {
		to = *datekvs.NewDateFromTime(*request.Params.To)
	}
	if to.Before(from.Time) {
		return nil, util.BadInputError(fmt.Errorf("to date is before from date"))
	}
	return &streamingGetStatsResponseObject{
		stream: shpanstream.MapStream[shpankids.TaskStats, openapi.ApiTaskStats](
			oa.taskManager.GetTaskStats(ctx, from, to),
			func(s *shpankids.TaskStats) *openapi.ApiTaskStats {
				return &openapi.ApiTaskStats{
					UserId:          s.UserId,
					ForDate:         s.ForDate,
					TotalTasksCount: s.TotalTasksCount,
					DoneTasksCount:  s.DoneTasksCount,
				}

			}),
		ctx: ctx,
	}, nil
}

type streamingGetStatsResponseObject struct {
	stream shpanstream.Stream[openapi.ApiTaskStats]
	ctx    context.Context
}

func (s *streamingGetStatsResponseObject) VisitGetStatsResponse(w http.ResponseWriter) error {
	return shpanstream.StreamToJsonResponseWriter(s.ctx, w, s.stream)
}

func NewOapiServerApiImpl(
	userSessionManager shpankids.UserSessionManager,
	userManager shpankids.UserManager,
	taskManager shpankids.TaskManager,
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
