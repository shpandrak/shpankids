package api

import (
	"context"
	"fmt"
	openapitypes "github.com/oapi-codegen/runtime/types"
	"shpankids/infra/util/castutil"
	"shpankids/internal/infra/util"
	"shpankids/openapi"
)

func (oa *OapiServerApiImpl) GetUserInfo(
	ctx context.Context,
	_ openapi.GetUserInfoRequestObject,
) (openapi.GetUserInfoResponseObject, error) {
	email, err := oa.userSessionManager(ctx)
	if err != nil {
		return nil, err
	}

	user, err := oa.userManager.FindUser(ctx, *email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, util.ForbiddenError(fmt.Errorf("%s is not a valid user", *email))
	}

	return openapi.GetUserInfo200JSONResponse{
		Email:     openapitypes.Email(*email),
		FirstName: castutil.ValToValPtr(user.FirstName),
		LastName:  castutil.ValToValPtr(user.LastName),
	}, nil
}
