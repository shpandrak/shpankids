package assignment

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"shpankids/infra/database/datekvs"
	"shpankids/infra/database/kvstore"
	"shpankids/infra/shpanstream"
	"shpankids/infra/util/functional"
	"shpankids/internal/infra/util"
	"shpankids/shpankids"
	"slices"
	"time"
)

type managerImpl struct {
	fs                 *firestore.Client
	kvs                kvstore.RawJsonStore
	userSessionManager shpankids.UserSessionManager
	sessionManager     shpankids.SessionManager
	familyManager      shpankids.FamilyManager
}

func NewAssignmentManager(
	kvs kvstore.RawJsonStore,
	userSessionManager shpankids.UserSessionManager,
	familyManager shpankids.FamilyManager,
	sessionManager shpankids.SessionManager,
) shpankids.AssignmentManager {
	return &managerImpl{
		kvs:                kvs,
		userSessionManager: userSessionManager,
		familyManager:      familyManager,
		sessionManager:     sessionManager,
	}
}

func (m *managerImpl) ListMyAssignmentsForToday(ctx context.Context) shpanstream.Stream[shpankids.DailyAssignmentDto] {
	userId, err := m.userSessionManager(ctx)
	if err != nil {
		return shpanstream.NewErrorStream[shpankids.DailyAssignmentDto](err)
	}

	s, err := m.sessionManager.Get(ctx, *userId)
	if err != nil {
		return shpanstream.NewErrorStream[shpankids.DailyAssignmentDto](err)
	}

	return m.doListAssignments(ctx, datekvs.TodayDate(s.Location), s.FamilyId, *userId)
}

func (m *managerImpl) GetAssignmentStats(
	ctx context.Context,
	fromDate datekvs.Date,
	toDate datekvs.Date,
) shpanstream.Stream[shpankids.AssignmentStats] {

	userId, err := m.userSessionManager(ctx)
	if err != nil {
		return shpanstream.NewErrorStream[shpankids.AssignmentStats](err)
	}
	s, err := m.sessionManager.Get(ctx, *userId)
	if err != nil {
		return shpanstream.NewErrorStream[shpankids.AssignmentStats](err)
	}

	f, err := m.familyManager.GetFamily(ctx, s.FamilyId)
	if err != nil {
		return shpanstream.NewErrorStream[shpankids.AssignmentStats](err)
	}

	isAdmin := slices.ContainsFunc(f.Members, func(fm shpankids.FamilyMemberDto) bool {
		return *userId == fm.UserId && fm.Role == shpankids.RoleAdmin
	})

	// Fetch only the tasks that are relevant to the user
	//(if the user is not an admin only fetch the tasks that are assigned to the user)
	userIdsToFetch := functional.MapSliceWhileFilteringNoErr(f.Members, func(member shpankids.FamilyMemberDto) *string {
		if isAdmin || *userId == member.UserId {
			return &member.UserId
		}
		return nil
	})

	// Get all the family tasks
	familyTasks, err := m.familyManager.ListFamilyTasks(ctx, s.FamilyId).CollectFilterNil(ctx)
	if err != nil {
		return shpanstream.NewErrorStream[shpankids.AssignmentStats](err)
	}

	return shpanstream.ConcatenatedStream[shpankids.AssignmentStats](
		functional.MapSliceNoErr(userIdsToFetch, func(userId string) shpanstream.Stream[shpankids.AssignmentStats] {
			return m.getUserTaskStatesForDateRange(
				ctx,
				fromDate,
				toDate,
				familyTasks,
				userId,
			)
		})...,
	)
}

func (m *managerImpl) doListAssignments(
	ctx context.Context,
	forDate datekvs.Date,
	familyId string,
	userId string,
) shpanstream.Stream[shpankids.DailyAssignmentDto] {
	familyTasks, err := m.familyManager.ListFamilyTasks(ctx, familyId).CollectFilterNil(ctx)
	if err != nil {
		return shpanstream.NewErrorStream[shpankids.DailyAssignmentDto](err)
	}

	// Combine different types of assignments into a single stream
	return shpanstream.ConcatenatedStream(
		m.filterTaskAssignmentsForUser(ctx, forDate, familyTasks, userId),
		m.filterProblemAssignmentsForUser(ctx, forDate, userId),
	)
}

func (m *managerImpl) filterProblemAssignmentsForUser(
	ctx context.Context,
	forDate datekvs.Date,
	userId string,
) shpanstream.Stream[shpankids.DailyAssignmentDto] {
	// todo:amit: consider loading ps managers externally maybe for the whole family at once to avoid the extra checks.
	// or cache family in ctx etc...
	userPsManager, err := m.familyManager.GetProblemSetManagerForUser(ctx, userId)
	if err != nil {
		return shpanstream.NewErrorStream[shpankids.DailyAssignmentDto](err)
	}

	return shpanstream.MapStreamWhileFilteringWithError(
		userPsManager.ListProblemSets(
			ctx,
		),
		func(ctx context.Context, fps *shpankids.ProblemSetDto) (*shpankids.DailyAssignmentDto, error) {
			return m.mapProblemSetToAssignment(ctx, userPsManager, fps, forDate)
		},
	)
}

func (m *managerImpl) mapProblemSetToAssignment(
	ctx context.Context,
	userPsManager shpankids.ProblemSetManager,
	fps *shpankids.ProblemSetDto,
	forDate datekvs.Date,
) (*shpankids.DailyAssignmentDto, error) {

	count, err := userPsManager.ListProblemSetSolutionsForDate(
		ctx,
		fps.ProblemSetId,
		forDate,
	).Count(ctx)
	if err != nil {
		return nil, err
	}

	status := shpankids.StatusOpen
	// If there are any solutions for the problem set, the status is set to done
	if count > 0 {
		status = shpankids.StatusDone
	} else {
		// Checking if there are no available problems for the problem set we're also done...
		noProblemsLeft, err :=
			userPsManager.ListProblemsForProblemSet(
				ctx,
				fps.ProblemSetId,
				false,
			).FindFirst().IsEmpty(ctx)

		if err != nil {
			return nil, err
		}
		if noProblemsLeft {
			// returning nil to filter away this assignment, no problems available...
			return nil, nil
		}
	}
	return &shpankids.DailyAssignmentDto{
		Id:          fps.ProblemSetId,
		ForDate:     forDate,
		Type:        shpankids.AssignmentTypeProblemSet,
		Title:       fps.Title,
		Status:      status,
		Description: fps.Description,
	}, nil
}

func (m *managerImpl) filterTaskAssignmentsForUser(
	ctx context.Context,
	forDate datekvs.Date,
	familyTasks []shpankids.FamilyTaskDto,
	userId string,
) shpanstream.Stream[shpankids.DailyAssignmentDto] {
	tasks := functional.MapSliceWhileFilteringNoErr(familyTasks, func(ft shpankids.FamilyTaskDto) **shpankids.DailyAssignmentDto {

		// Filtering away tasks that were deleted after the forDate
		if ft.Status != shpankids.FamilyAssignmentStatusActive && ft.StatusDate.Before(forDate.Time) {
			return nil
		}

		// Filtering away tasks that are not assigned to the user
		if !slices.ContainsFunc(ft.MemberIds, func(memberId string) bool {
			return memberId == userId
		}) {
			return nil
		}
		return functional.ValueToPointer(&shpankids.DailyAssignmentDto{
			Id:          ft.TaskId,
			ForDate:     forDate,
			Type:        shpankids.AssignmentTypeTask,
			Title:       ft.Title,
			Status:      shpankids.StatusOpen,
			Description: ft.Description,
		})
	})

	assignmentsById := functional.SliceToMapNoErr(tasks, func(t *shpankids.DailyAssignmentDto) string {
		return t.Id
	})

	userTaskRepo, err := NewUserTaskStatusRepository(ctx, m.kvs, userId)
	if err != nil {
		return shpanstream.NewErrorStream[shpankids.DailyAssignmentDto](err)
	}

	err = userTaskRepo.StreamAllForDate(ctx, forDate).Consume(
		ctx,
		func(dr *functional.Entry[string, dbUserTaskStatus]) {
			foundTask, found := assignmentsById[dr.Key]
			if found {
				foundTask.Status = dr.Value.Status
			}
		},
	)
	if err != nil {
		return shpanstream.NewErrorStream[shpankids.DailyAssignmentDto](err)
	}

	return shpanstream.Just[shpankids.DailyAssignmentDto](functional.MapSliceUnPtr(functional.MapValues(assignmentsById))...)
}

func (m *managerImpl) getUserTaskStatesForDateRange(
	ctx context.Context,
	from datekvs.Date,
	to datekvs.Date,
	familyTasks []shpankids.FamilyTaskDto,
	userId string,
) shpanstream.Stream[shpankids.AssignmentStats] {

	relevantUserTasks := functional.FilterSlice(familyTasks, func(ft shpankids.FamilyTaskDto) bool {
		// Filtering away tasks that were deleted after the forDate
		if ft.Status != shpankids.FamilyAssignmentStatusActive && ft.StatusDate.Before(from.Time) {
			return false
		}

		if ft.Created.After(to.DateEndTime()) {
			return false
		}

		// Filtering away tasks that are not assigned to the user
		if !slices.ContainsFunc(ft.MemberIds, func(memberId string) bool {
			return memberId == userId
		}) {
			return false
		}
		return true
	})

	userTaskRepo, err := NewUserTaskStatusRepository(ctx, m.kvs, userId)
	if err != nil {
		return shpanstream.NewErrorStream[shpankids.AssignmentStats](err)
	}

	return shpanstream.MapStreamWhileFilteringWithError[datekvs.Date, shpankids.AssignmentStats](
		datekvs.NewDateRangeStream(from, to),
		func(ctx context.Context, dt *datekvs.Date) (*shpankids.AssignmentStats, error) {
			userAssignableTasksByTaskId := functional.SliceToMapNoErr(
				functional.FilterSlice(relevantUserTasks, func(t shpankids.FamilyTaskDto) bool {
					return t.Created.Before(dt.DateEndTime()) &&
						(t.Status == shpankids.FamilyAssignmentStatusActive || t.StatusDate.After(dt.Time))
				}),
				func(t shpankids.FamilyTaskDto) string {
					return t.TaskId
				},
			)

			if len(userAssignableTasksByTaskId) == 0 {
				return nil, nil
			}

			ret := &shpankids.AssignmentStats{
				UserId:          userId,
				ForDate:         dt.Time,
				TotalTasksCount: len(userAssignableTasksByTaskId),
				DoneTasksCount:  0,
			}

			err := userTaskRepo.StreamAllForDate(ctx, *dt).Consume(
				ctx,
				func(dr *functional.Entry[string, dbUserTaskStatus]) {

					if _, found := userAssignableTasksByTaskId[dr.Key]; found {
						if dr.Value.Status == shpankids.StatusDone {
							ret.DoneTasksCount++
						}
					}
				},
			)
			if err != nil {
				return nil, err
			}
			return ret, nil
		})

}

func (m *managerImpl) UpdateAssignmentStatus(
	ctx context.Context,
	forDay time.Time,
	taskId string,
	status shpankids.AssignmentStatus,
	comment string,
) error {
	userId, err := m.userSessionManager(ctx)
	if err != nil {
		return err
	}

	tsr, err := NewUserTaskStatusRepository(ctx, m.kvs, *userId)
	if err != nil {
		return err
	}

	return tsr.Set(ctx, *datekvs.NewDateFromTime(forDay), taskId, dbUserTaskStatus{
		Comment:    comment,
		Status:     status,
		StatusTime: time.Now(),
	})

}

func (m *managerImpl) CreateNewAssignment(
	ctx context.Context,
	forUserId string,
	args shpankids.CreateAssignmentArgsDto,
) error {
	if args.Id == "" {
		return util.BadInputError(fmt.Errorf("id is required"))
	}
	loggedInUserId, err := m.userSessionManager(ctx)
	if err != nil {
		return err
	}

	userAssignmentsRepo, err := newUserAssignmentsRepository(ctx, m.kvs, forUserId)
	if err != nil {
		return err
	}

	return userAssignmentsRepo.Set(
		ctx,
		args.Id,
		dbUserAssignment{
			AssignmentType:  args.Type,
			Title:           args.Title,
			Created:         time.Now(),
			NumberOfParts:   args.NumberOfParts,
			CreatedByUserId: *loggedInUserId,
			Description:     args.Description,
		},
	)
}

func (m *managerImpl) ArchiveUserAssignment(ctx context.Context, forUserId string, assignmentId string) error {

	userAssignmentsRepo, err := newUserAssignmentsRepository(ctx, m.kvs, forUserId)
	if err != nil {
		return err
	}

	return userAssignmentsRepo.Archive(ctx, assignmentId)
}

func (m *managerImpl) ReportTaskProgress(
	ctx context.Context,
	forUserId string,
	forDate datekvs.Date,
	assignmentId string,
	partsDelta int,
	comment string,
) error {
	if partsDelta == 0 {
		return nil
	}
	// todo:amit: check whether this is allowed (e.g. on behalf? or is this done elsewhere?)
	userAssignmentsRepo, err := newAssignmentStatusRepo(ctx, m.kvs, forUserId)
	if err != nil {
		return err
	}

	return userAssignmentsRepo.ManipulateOrCreate(ctx, forDate, assignmentId, func(es *dbAssignmentStatus) (dbAssignmentStatus, error) {

		// for first time, create the assignment status, but it has to be possible to update it
		if es == nil {
			if partsDelta < 0 {
				return functional.DefaultValue[dbAssignmentStatus](), util.BadInputError(fmt.Errorf(
					"cannot decrease parts for assignment %s that has not yet started progress",
					assignmentId,
				))
			}
			newComments := es.Comments
			if comment != "" {
				newComments = append(es.Comments, comment)
			}
			return dbAssignmentStatus{
				Comments:       newComments,
				PartsCompleted: partsDelta,
				LastUpdated:    time.Now(),
			}, nil
		} else {
			newPatsCompleted := es.PartsCompleted + partsDelta
			if newPatsCompleted < 0 {
				return functional.DefaultValue[dbAssignmentStatus](), util.BadInputError(fmt.Errorf(
					"cannot decrease parts completed for assignment %s below 0. already completed %d, tried delta %d",
					assignmentId,
					es.PartsCompleted,
					partsDelta,
				))
			}
			newComments := es.Comments
			if comment != "" {
				newComments = append(es.Comments, comment)
			}
			return dbAssignmentStatus{
				Comments:       newComments,
				PartsCompleted: newPatsCompleted,
				LastUpdated:    time.Now(),
			}, nil
		}
	})

}
