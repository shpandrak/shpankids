package api

import (
	"context"
	"shpankids/infra/util/castutil"
	"shpankids/openapi"
	"shpankids/shpankids"
)

func (oa *OapiServerApiImpl) UpdateTaskStatus(
	ctx context.Context,
	request openapi.UpdateTaskStatusRequestObject,
) (openapi.UpdateTaskStatusResponseObject, error) {
	err := oa.taskManager.UpdateTaskStatus(
		ctx,
		request.Body.ForDate,
		request.Body.TaskId,
		shpankids.Status(request.Body.Status),
		castutil.ValPtrToVal(request.Body.Comment),
	)
	if err != nil {
		return nil, err
	}
	return openapi.UpdateTaskStatus200Response{}, nil
}
