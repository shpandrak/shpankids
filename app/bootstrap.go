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
const alexUserId = "alex.lieberman.matu@gmail.com"
const charlotteUserId = "alma.lieberman@gmail.com"
const nemalaUserId = "yael.peer@gmail.com"

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
		{
			Email:     alexUserId,
			FirstName: "Alex",
			LastName:  "Lieberman",
			BirthDate: time.Date(2016, 9, 7, 0, 0, 0, 0, time.UTC),
		},
		{
			Email:     charlotteUserId,
			FirstName: "Alma Charlotte",
			LastName:  "Lieberman",
			BirthDate: time.Date(2012, 8, 21, 0, 0, 0, 0, time.UTC),
		},
		{
			Email:     nemalaUserId,
			FirstName: "Yael",
			LastName:  "Lieberman",
			BirthDate: time.Date(1982, 6, 15, 0, 0, 0, 0, time.UTC),
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
				charlotteUserId,
				alexUserId,
			},
			[]string{
				shpanUserId,
				nemalaUserId,
			},
		)
		if err != nil {
			return err
		}

		defaultFamilyTasks := []shpankids.FamilyTaskDto{
			{
				TaskId:      "task1",
				Title:       "להכין שעורי בית",
				Description: "Do your homework",
				MemberIds:   []string{peteUserId, alexUserId, charlotteUserId},
			},
			{
				TaskId:      "task2",
				Title:       "להוציא קופסאת אוכל",
				Description: "Lunch box -> dishwasher",
				MemberIds:   []string{peteUserId, alexUserId, charlotteUserId},
			},
			{
				TaskId:    "task3",
				Title:     "לנגן",
				MemberIds: []string{peteUserId},
			},
			{
				TaskId:    "task4",
				Title:     "דואולינגו",
				MemberIds: []string{charlotteUserId},
			},
		}

		for _, currTask := range defaultFamilyTasks {
			// Create default tasks for the family
			err = familyManager.CreateFamilyTask(
				bootstrapCtx,
				shpanFamilyId,
				currTask,
			)
			if err != nil {
				return err
			}
		}

		problemSet1Id := "problemSet1"
		err = familyManager.CreateFamilyProblemSet(
			bootstrapCtx,
			shpanFamilyId,
			peteUserId,
			shpankids.FamilyProblemSetDto{
				ProblemSetId: problemSet1Id,
				Title:        "שאלות מוזיקה",
			},
		)
		if err != nil {
			return err
		}

		problemsPack := []shpankids.FamilyProblemDto{
			{
				ProblemId: "problem1",
				Title:     "מהם הצלילים באקורד רה מינור?",
				Alternatives: []shpankids.ProblemAlternativeDto{
					{
						Title: "רה-סול-דו",
					},
					{
						Title: "רה-סול-סי",
					},
					{
						Title:   "רה-פה-לה",
						Correct: true,
					},
					{
						Title: "רה-פה#-לה",
					},
				},
			},
			{
				ProblemId: "problem2",
				Title:     "מהו הצליל השישי בסולם דו מז׳ור?",
				Alternatives: []shpankids.ProblemAlternativeDto{
					{
						Title: "רה",
					},
					{
						Title:   "לה",
						Correct: true,
					},
					{
						Title: "מי במול",
					},
					{
						Title: "פה",
					},
				},
			},
		}
		for _, currProblem := range problemsPack {
			err = familyManager.CreateFamilyProblem(bootstrapCtx, shpanFamilyId, peteUserId, problemSet1Id, currProblem)
			if err != nil {
				return err
			}
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
			ourLocation, err := time.LoadLocation("Asia/Jerusalem")
			if err != nil {
				return err
			}
			err = sessionManager.Set(bootstrapCtx, currDefaultUser.Email, shpankids.Session{
				FamilyId: shpanFamilyId,
				Location: ourLocation,
			})

		}

	}
	return nil
}
