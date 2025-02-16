package api

import (
	"context"
	"fmt"
	"net/http"
	"shpankids/infra/database/datekvs"
	"shpankids/infra/shpanstream"
	"shpankids/infra/util/castutil"
	"shpankids/infra/util/functional"
	"shpankids/internal/infra/util"
	"shpankids/openapi"
	"shpankids/shpankids"
	"time"
)

type OapiServerApiImpl struct {
	userSessionManager shpankids.UserSessionManager
	userManager        shpankids.UserManager
	assignmentManager  shpankids.AssignmentManager
	familyManager      shpankids.FamilyManager
	sessionManager     shpankids.SessionManager
}

func (oa *OapiServerApiImpl) GetStats(
	ctx context.Context,
	request openapi.GetStatsRequestObject,
) (openapi.GetStatsResponseObject, error) {

	from := *datekvs.NewDateFromTime(time.Now())
	to := from

	userId, err := oa.userSessionManager(ctx)
	if err != nil {
		return nil, err
	}

	// todo:add put session in context, avoid accessing multiple times
	s, err := oa.sessionManager.Get(ctx, *userId)
	if err != nil {
		return nil, err
	}

	if request.Params.From != nil {
		from = *datekvs.NewDateFromTime(request.Params.From.In(s.Location))
	}
	if request.Params.To != nil {
		to = *datekvs.NewDateFromTime(request.Params.To.In(s.Location))
	}

	if to.Before(from.Time) {
		return nil, util.BadInputError(fmt.Errorf("to date is before from date"))
	}
	return &streamingGetStatsResponseObject{
		stream: shpanstream.MapStream[shpankids.TaskStats, openapi.ApiTaskStats](
			oa.assignmentManager.GetTaskStats(ctx, from, to),
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
	assignmentManager shpankids.AssignmentManager,
	familyManager shpankids.FamilyManager,
	sessionManager shpankids.SessionManager,

) *OapiServerApiImpl {
	return &OapiServerApiImpl{
		userSessionManager: userSessionManager,
		userManager:        userManager,
		assignmentManager:  assignmentManager,
		familyManager:      familyManager,
		sessionManager:     sessionManager,
	}
}

func (oa *OapiServerApiImpl) ListAssignments(
	ctx context.Context,
	_ openapi.ListAssignmentsRequestObject,
) (openapi.ListAssignmentsResponseObject, error) {
	AssignmentList, err := oa.assignmentManager.ListAssignmentsForToday(ctx)
	if err != nil {
		return nil, err
	}
	return openapi.ListAssignments200JSONResponse(functional.MapSliceNoErr(AssignmentList, toUiAssignment)), nil
}

func toUiAssignment(a shpankids.Assignment) openapi.ApiAssignment {
	return openapi.ApiAssignment{
		Description: castutil.ValToValPtr(a.Description),
		ForDate:     a.ForDate.Time,
		Id:          a.Id,
		Status:      openapi.ApiAssignmentStatus(a.Status),
		Title:       a.Title,
		Type:        openapi.ApiAssignmentType(a.Type),
	}
}
