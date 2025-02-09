package api

import (
	"context"
	"shpankids/domain/task"
	"shpankids/infra/util/castutil"
	"shpankids/openapi"
)

func (oa *OapiServerApiImpl) UpdateTaskStatus(
	ctx context.Context,
	request openapi.UpdateTaskStatusRequestObject,
) (openapi.UpdateTaskStatusResponseObject, error) {
	err := oa.taskManager.UpdateTaskStatus(
		ctx,
		request.Body.ForDate,
		request.Body.TaskId,
		task.Status(request.Body.Status),
		castutil.ValPtrToVal(request.Body.Comment),
	)
	if err != nil {
		return nil, err
	}
	return openapi.UpdateTaskStatus200Response{}, nil
}
