package api

import (
	"context"
	"fmt"
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

func (oa *OapiServerApiImpl) GetProblem(ctx context.Context, request openapi.GetProblemRequestObject) (openapi.GetProblemResponseObject, error) {
	_, s, err := oa.getUserAndSession(ctx)
	if err != nil {
		return nil, err
	}
	p, err := oa.familyManager.GetProblem(ctx, s.FamilyId, request.UserId, request.ProblemSetId, request.ProblemId)
	if err != nil {
		return nil, err
	}

	apiPrb, err := toApiProblem(ctx, p)
	if err != nil {
		return nil, err
	}
	return openapi.GetProblem200JSONResponse(*apiPrb), nil
}

func (oa *OapiServerApiImpl) ListUserProblemsSolutions(
	ctx context.Context,
	request openapi.ListUserProblemsSolutionsRequestObject,
) (openapi.ListUserProblemsSolutionsResponseObject, error) {

	_, s, err := oa.getUserAndSession(ctx)
	if err != nil {
		return nil, err
	}
	return &streamingProblemSetSolution{
		ctx: ctx,
		stream: oa.familyManager.ListUserProblemsSolutions(
			ctx,
			s.FamilyId,
			request.ProblemSetId,
			request.UserId,
		),
	}, nil
}

func (oa *OapiServerApiImpl) SubmitProblemAnswer(
	ctx context.Context,
	request openapi.SubmitProblemAnswerRequestObject,
) (openapi.SubmitProblemAnswerResponseObject, error) {
	userId, s, err := oa.getUserAndSession(ctx)
	if err != nil {
		return nil, err
	}
	isCorrect, correctAnswerId, problemDto, err := oa.familyManager.SubmitProblemAnswer(
		ctx,
		s.FamilyId,
		userId,
		request.Body.AssignmentId,
		request.Body.ProblemId,
		datekvs.TodayDate(s.Location),
		request.Body.AnswerId,
	)
	if err != nil {
		return nil, err
	}
	return &openapi.SubmitProblemAnswer200JSONResponse{
		CorrectAnswerId: correctAnswerId,
		Explanation:     castutil.StrToStrPtr(problemDto.Explanation),
		IsCorrect:       isCorrect,
	}, nil
}

func (oa *OapiServerApiImpl) CreateProblemsInSet(
	ctx context.Context,
	request openapi.CreateProblemsInSetRequestObject,
) (openapi.CreateProblemsInSetResponseObject, error) {
	_, s, err := oa.getUserAndSession(ctx)
	if err != nil {
		return nil, err
	}
	err = oa.familyManager.CreateProblemsInSet(
		ctx,
		s.FamilyId,
		request.Body.ForUserId,
		request.Body.ProblemSetId,
		functional.MapSliceNoErr(request.Body.Problems, toCreateFamilyProblemDto),
	)
	if err != nil {
		return nil, err
	}
	return &openapi.CreateProblemsInSet200Response{}, nil

}

func toCreateFamilyProblemDto(p openapi.ApiProblemForEdit) shpankids.CreateProblemDto {
	return shpankids.CreateProblemDto{
		Description: castutil.StrPtrToStr(p.Description),
		Title:       p.Title,
		Answers:     functional.MapSliceNoErr(p.Answers, toCreateProblemAnswerDto),
	}
}

func toCreateProblemAnswerDto(a openapi.ApiProblemAnswerForEdit) shpankids.CreateProblemAnswerDto {
	return shpankids.CreateProblemAnswerDto{
		Description: castutil.StrPtrToStr(a.Description),
		Title:       a.Title,
		Correct:     a.IsCorrect,
	}
}

func (oa *OapiServerApiImpl) getUserAndSession(ctx context.Context) (string, *shpankids.Session, error) {
	// todo:add put session in context, avoid accessing multiple times
	userId, err := oa.userSessionManager(ctx)
	if err != nil {
		return "", nil, err
	}
	s, err := oa.sessionManager.Get(ctx, *userId)
	if err != nil {
		return "", nil, err
	}
	return *userId, s, nil
}

func (oa *OapiServerApiImpl) ListProblemSetProblems(
	ctx context.Context,
	request openapi.ListProblemSetProblemsRequestObject,
) (openapi.ListProblemSetProblemsResponseObject, error) {
	_, s, err := oa.getUserAndSession(ctx)
	if err != nil {
		return nil, err
	}
	return &streamingProblemsForEdit{
		ctx: ctx,
		stream: shpanstream.MapStream(
			oa.familyManager.ListProblemsForProblemSet(
				ctx,
				s.FamilyId,
				request.UserId,
				request.ProblemSetId,
				false,
			),
			ToApiProblemForEdit,
		),
	}, nil
}

func ToApiProblemForEdit(p *shpankids.FamilyProblemDto) *openapi.ApiProblemForEdit {
	return &openapi.ApiProblemForEdit{
		Description: castutil.StrToStrPtr(p.Description),
		Id:          functional.ValueToPointer(p.ProblemId),
		Title:       p.Title,
		Answers:     functional.MapSliceNoErr(p.Answers, toApiProblemAnswerForEdit),
	}

}
func toApiProblemAnswerForEdit(a shpankids.ProblemAnswerDto) openapi.ApiProblemAnswerForEdit {
	return openapi.ApiProblemAnswerForEdit{
		Description: castutil.StrToStrPtr(a.Description),
		Id:          functional.ValueToPointer(a.Id),
		IsCorrect:   a.Correct,
		Title:       a.Title,
	}
}

func (oa *OapiServerApiImpl) ListUserFamilyProblemSets(
	ctx context.Context,
	request openapi.ListUserFamilyProblemSetsRequestObject,
) (openapi.ListUserFamilyProblemSetsResponseObject, error) {
	_, s, err := oa.getUserAndSession(ctx)
	if err != nil {
		return nil, err
	}
	return &streamingProblemSets{
		stream: shpanstream.MapStream(
			oa.familyManager.ListProblemSetsForUser(ctx, s.FamilyId, request.Params.UserId),
			toApiProblemSet,
		),
		ctx: ctx,
	}, nil
}
func toApiProblemSet(p *shpankids.FamilyProblemSetDto) *openapi.ApiProblemSet {
	return &openapi.ApiProblemSet{
		Id:          p.ProblemSetId,
		Title:       p.Title,
		Description: castutil.StrToStrPtr(p.Description),
	}
}

func (oa *OapiServerApiImpl) LoadProblemForAssignment(
	ctx context.Context,
	request openapi.LoadProblemForAssignmentRequestObject,
) (openapi.LoadProblemForAssignmentResponseObject, error) {
	userId, s, err := oa.getUserAndSession(ctx)
	if err != nil {
		return nil, err
	}

	first, err := shpanstream.MapStreamWithError(
		oa.familyManager.ListProblemsForProblemSet(ctx, s.FamilyId, userId, request.Body.AssignmentId, false),
		toApiProblem,
	).GetFirst(ctx)

	if err != nil {
		return nil, err
	}
	if first == nil {
		return nil, util.NotFoundError(fmt.Errorf("no problems found for set"))
	} else {
		return &openapi.LoadProblemForAssignment200JSONResponse{Problem: *first}, nil
	}

}

func toApiProblem(_ context.Context, p *shpankids.FamilyProblemDto) (*openapi.ApiProblem, error) {
	mapAnswers, err := functional.MapSliceWithIdx(p.Answers, func(idx int, a shpankids.ProblemAnswerDto) (openapi.ApiProblemAnswer, error) {
		return openapi.ApiProblemAnswer{
			Id:    a.Id,
			Title: a.Title,
		}, nil
	})
	if err != nil {
		return nil, err
	}
	return &openapi.ApiProblem{
		Description: castutil.StrToStrPtr(p.Description),
		Id:          p.ProblemId,
		Title:       p.Title,
		Answers:     mapAnswers,
	}, nil
}

func (oa *OapiServerApiImpl) GetStats(
	ctx context.Context,
	request openapi.GetStatsRequestObject,
) (openapi.GetStatsResponseObject, error) {

	from := *datekvs.NewDateFromTime(time.Now())
	to := from

	_, s, err := oa.getUserAndSession(ctx)
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
	return &streamingAssignments{
		ctx: ctx,
		stream: shpanstream.MapStream[shpankids.Assignment, openapi.ApiAssignment](
			oa.assignmentManager.ListAssignmentsForToday(ctx),
			toApiAssignment,
		),
	}, nil

}

func toApiAssignment(a *shpankids.Assignment) *openapi.ApiAssignment {
	return &openapi.ApiAssignment{
		Description: castutil.ValToValPtr(a.Description),
		ForDate:     a.ForDate.Time,
		Id:          a.Id,
		Status:      openapi.ApiAssignmentStatus(a.Status),
		Title:       a.Title,
		Type:        openapi.ApiAssignmentType(a.Type),
	}
}
