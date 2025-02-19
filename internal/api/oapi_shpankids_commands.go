package api

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"shpankids/infra/util/castutil"
	"shpankids/openapi"
	"shpankids/shpankids"
	"time"
)

func (oa *OapiServerApiImpl) UpdateTaskStatus(
	ctx context.Context,
	request openapi.UpdateTaskStatusRequestObject,
) (openapi.UpdateTaskStatusResponseObject, error) {
	err := oa.assignmentManager.UpdateTaskStatus(
		ctx,
		request.Body.ForDate,
		request.Body.TaskId,
		shpankids.AssignmentStatus(request.Body.Status),
		castutil.ValPtrToVal(request.Body.Comment),
	)
	if err != nil {
		return nil, err
	}
	return openapi.UpdateTaskStatus200Response{}, nil
}

func (oa *OapiServerApiImpl) CreateFamilyTask(ctx context.Context, request openapi.CreateFamilyTaskRequestObject) (openapi.CreateFamilyTaskResponseObject, error) {
	uId, err := oa.userSessionManager(ctx)
	if err != nil {
		return nil, err
	}
	s, err := oa.sessionManager.Get(ctx, *uId)
	if err != nil {
		return nil, err
	}
	err = oa.familyManager.CreateFamilyTask(ctx, s.FamilyId, shpankids.FamilyTaskDto{
		TaskId:      uuid.NewString(),
		Title:       request.Body.Task.Title,
		Description: castutil.StrPtrToStr(request.Body.Task.Description),
		MemberIds:   request.Body.Task.MemberIds,
		Created:     time.Now(),
	})
	if err != nil {
		return nil, err
	}
	return openapi.CreateFamilyTask200Response{}, nil
}

func (oa *OapiServerApiImpl) UpdateFamilyTask(ctx context.Context, request openapi.UpdateFamilyTaskRequestObject) (openapi.UpdateFamilyTaskResponseObject, error) {
	return nil, fmt.Errorf("not implemented")
}
func (oa *OapiServerApiImpl) DeleteFamilyTask(
	ctx context.Context,
	request openapi.DeleteFamilyTaskRequestObject,
) (openapi.DeleteFamilyTaskResponseObject, error) {
	uId, err := oa.userSessionManager(ctx)
	if err != nil {
		return nil, err
	}
	s, err := oa.sessionManager.Get(ctx, *uId)
	if err != nil {
		return nil, err
	}
	err = oa.familyManager.DeleteFamilyTask(ctx, s.FamilyId, request.Body.TaskId)
	if err != nil {
		return nil, err
	}
	return openapi.DeleteFamilyTask200Response{}, nil
}

func (oa *OapiServerApiImpl) CreateProblemSet(
	ctx context.Context,
	request openapi.CreateProblemSetRequestObject,
) (openapi.CreateProblemSetResponseObject, error) {
	_, s, err := oa.getUserAndSession(ctx)
	if err != nil {
		return nil, err
	}
	err = oa.familyManager.CreateProblemSet(
		ctx,
		s.FamilyId,
		request.Body.ForUserId,
		shpankids.CreateProblemSetDto{
			ProblemSetId: uuid.NewString(),
			Title:        request.Body.Title,
			Description:  castutil.StrPtrToStr(request.Body.Description),
		})
	if err != nil {
		return nil, err
	}
	return openapi.CreateProblemSet200Response{}, nil
}
