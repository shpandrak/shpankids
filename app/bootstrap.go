package app

import (
	"context"
	"shpankids/shpankids"
	"shpankids/webserver/auth"
	"time"
)

const shpanUserId = "shpandrak@gmail.com"
const shpanFamilyId = "shpanFamily"
const peteUserId = "pete.lieberman.real@gmail.com"

func appBootstrap(
	userManager shpankids.UserManager,
	familyManager shpankids.FamilyManager,
	sessionManager shpankids.SessionManager,
) error {

	defaultUsers := []shpankids.User{
		{
			Email:     shpanUserId,
			FirstName: "Amit",
			LastName:  "Lieberman",
			BirthDate: time.Date(1981, 3, 5, 13, 30, 0, 0, time.UTC),
		},
		{
			Email:     peteUserId,
			FirstName: "Pete",
			LastName:  "Lieberman",
			BirthDate: time.Date(2016, 9, 7, 0, 0, 0, 0, time.UTC),
		},
	}

	// Create the default family
	bootstrapCtx := auth.EnrichContext(context.Background(), shpanUserId)

	existingShpanFam, err := familyManager.FindFamily(bootstrapCtx, shpanFamilyId)
	if err != nil {
		return err
	}
	if existingShpanFam == nil {
		err = familyManager.CreateFamily(
			bootstrapCtx,
			shpanFamilyId,
			"Lieberman Family",
			[]string{
				peteUserId,
			},
		)
		if err != nil {
			return err
		}

		// Create default tasks for the family
		err = familyManager.CreateFamilyTask(
			bootstrapCtx,
			shpanFamilyId,
			shpankids.FamilyTaskDto{
				TaskId:      "task1",
				Title:       "Do Homework",
				Description: "Do your homework",
				MemberIds:   []string{peteUserId, shpanUserId},
			},
		)
		if err != nil {
			return err
		}
		err = familyManager.CreateFamilyTask(
			bootstrapCtx,
			shpanFamilyId,
			shpankids.FamilyTaskDto{
				TaskId:      "task2",
				Title:       "Take out your lunch box",
				Description: "Lunch box -> dishwasher",
				MemberIds:   []string{peteUserId, shpanUserId},
			},
		)
		if err != nil {
			return err
		}

	}

	for _, currDefaultUser := range defaultUsers {
		usr, err := userManager.FindUser(bootstrapCtx, currDefaultUser.Email)
		if err != nil {
			return err
		}
		if usr == nil {
			err = userManager.CreateUser(
				bootstrapCtx,
				currDefaultUser.Email,
				currDefaultUser.FirstName,
				currDefaultUser.LastName,
				currDefaultUser.BirthDate,
			)
			if err != nil {
				return err
			}

			// Create a session for the user
			err = sessionManager.Set(bootstrapCtx, currDefaultUser.Email, shpankids.Session{
				FamilyId: shpanFamilyId,
			})

		}

	}
	return nil
}
