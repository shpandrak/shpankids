package api

import (
	"context"
	"fmt"
	openapitypes "github.com/oapi-codegen/runtime/types"
	"shpankids/infra/util/castutil"
	"shpankids/infra/util/functional"
	"shpankids/internal/infra/util"
	"shpankids/openapi"
	"shpankids/shpankids"
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

func (oa *OapiServerApiImpl) GetFamilyInfo(
	ctx context.Context,
	_ openapi.GetFamilyInfoRequestObject,
) (openapi.GetFamilyInfoResponseObject, error) {
	usrId, err := oa.userSessionManager(ctx)
	if err != nil {
		return nil, err
	}

	s, err := oa.sessionManager.Get(ctx, *usrId)
	if err != nil {
		return nil, err
	}
	family, err := oa.familyManager.FindFamily(ctx, s.FamilyId)
	if err != nil {
		return nil, err
	}
	if family == nil {
		return nil, util.NotFoundError(fmt.Errorf("family %s not found", s.FamilyId))
	}

	uiFamilyMembers, err := functional.MapSlice(family.Members, func(member shpankids.FamilyMemberDto) (openapi.UIFamilyMember, error) {
		currUsr, err := oa.userManager.GetUser(ctx, member.UserId)

		if err != nil {
			return functional.DefaultValue[openapi.UIFamilyMember](), err
		}
		return openapi.UIFamilyMember{
			Email:     openapitypes.Email(currUsr.Email),
			FirstName: currUsr.FirstName,
			LastName:  currUsr.LastName,
			Role:      openapi.ApiFamilyRole(member.Role),
		}, nil

	})
	if err != nil {
		return nil, err
	}

	familyTasks, err := oa.familyManager.ListFamilyTasks(ctx, family.Id).CollectFilterNil(ctx)
	if err != nil {
		return nil, err
	}

	return openapi.GetFamilyInfo200JSONResponse{
		AdminEmail:        openapitypes.Email(family.OwnerEmail),
		FamilyDisplayName: family.Name,
		FamilyUri:         family.Id,
		Members:           uiFamilyMembers,
		Tasks: functional.MapSliceWhileFilteringNoErr(familyTasks, func(task shpankids.FamilyTaskDto) *openapi.UIFamilyTask {
			if task.Status == shpankids.FamilyTaskStatusActive {
				return &openapi.UIFamilyTask{
					Description: castutil.StrToStrPtr(task.Description),
					Id:          task.TaskId,
					MemberIds:   task.MemberIds,
					Title:       task.Title,
				}
			} else {
				return nil
			}
		}),
	}, nil
}
