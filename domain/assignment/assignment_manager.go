package assignment

import (
	"cloud.google.com/go/firestore"
	"context"
	"shpankids/infra/database/datekvs"
	"shpankids/infra/database/kvstore"
	"shpankids/infra/shpanstream"
	"shpankids/infra/util/functional"
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

func (m *managerImpl) ListAssignmentsForToday(ctx context.Context) shpanstream.Stream[shpankids.Assignment] {
	userId, err := m.userSessionManager(ctx)
	if err != nil {
		return shpanstream.NewErrorStream[shpankids.Assignment](err)
	}

	s, err := m.sessionManager.Get(ctx, *userId)
	if err != nil {
		return shpanstream.NewErrorStream[shpankids.Assignment](err)
	}

	return m.doListAssignments(ctx, datekvs.TodayDate(s.Location), s.FamilyId, *userId)
}

func NewAssignmentManager(
	fs *firestore.Client,
	kvs kvstore.RawJsonStore,
	userSessionManager shpankids.UserSessionManager,
	familyManager shpankids.FamilyManager,
	sessionManager shpankids.SessionManager,
) shpankids.AssignmentManager {
	return &managerImpl{
		fs:                 fs,
		kvs:                kvs,
		userSessionManager: userSessionManager,
		familyManager:      familyManager,
		sessionManager:     sessionManager,
	}
}

func (m *managerImpl) GetTaskStats(
	ctx context.Context,
	fromDate datekvs.Date,
	toDate datekvs.Date,
) shpanstream.Stream[shpankids.TaskStats] {

	userId, err := m.userSessionManager(ctx)
	if err != nil {
		return shpanstream.NewErrorStream[shpankids.TaskStats](err)
	}
	s, err := m.sessionManager.Get(ctx, *userId)
	if err != nil {
		return shpanstream.NewErrorStream[shpankids.TaskStats](err)
	}

	f, err := m.familyManager.GetFamily(ctx, s.FamilyId)
	if err != nil {
		return shpanstream.NewErrorStream[shpankids.TaskStats](err)
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

	familyTasks, err := m.familyManager.ListFamilyTasks(ctx, s.FamilyId).CollectFilterNil(ctx)
	if err != nil {
		return shpanstream.NewErrorStream[shpankids.TaskStats](err)
	}
	return shpanstream.ConcatenatedStream[shpankids.TaskStats](
		functional.MapSliceNoErr(userIdsToFetch, func(userId string) shpanstream.Stream[shpankids.TaskStats] {
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
) shpanstream.Stream[shpankids.Assignment] {
	familyTasks, err := m.familyManager.ListFamilyTasks(ctx, familyId).CollectFilterNil(ctx)
	if err != nil {
		return shpanstream.NewErrorStream[shpankids.Assignment](err)
	}

	// Combine different types of assignments into a single stream
	return shpanstream.ConcatenatedStream(
		m.filterTaskAssignmentsForUser(ctx, forDate, familyTasks, userId),
		m.filterProblemAssignmentsForUser(ctx, forDate, familyId, userId),
	)
}

func (m *managerImpl) filterProblemAssignmentsForUser(
	ctx context.Context,
	forDate datekvs.Date,
	familyId string,
	userId string,
) shpanstream.Stream[shpankids.Assignment] {
	return shpanstream.MapStreamWithError(
		m.familyManager.ListProblemSetsForUser(
			ctx,
			familyId,
			userId,
		),
		func(ctx context.Context, fps *shpankids.FamilyProblemSetDto) (*shpankids.Assignment, error) {
			return m.mapProblemSetToAssignment(ctx, familyId, userId, fps, forDate)
		},
	)
}

func (m *managerImpl) mapProblemSetToAssignment(
	ctx context.Context,
	familyId string,
	userId string,
	fps *shpankids.FamilyProblemSetDto,
	forDate datekvs.Date,
) (*shpankids.Assignment, error) {
	count, err := m.familyManager.ListProblemSetSolutionsForDate(
		ctx,
		familyId,
		userId,
		fps.ProblemSetId,
		forDate,
	).Count(ctx)
	if err != nil {
		return nil, err
	}

	status := shpankids.StatusOpen
	if count > 0 {
		status = shpankids.StatusDone
	}
	return &shpankids.Assignment{
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
) shpanstream.Stream[shpankids.Assignment] {
	tasks := functional.MapSliceWhileFilteringNoErr(familyTasks, func(ft shpankids.FamilyTaskDto) **shpankids.Assignment {

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
		return functional.ValueToPointer(&shpankids.Assignment{
			Id:          ft.TaskId,
			ForDate:     forDate,
			Type:        shpankids.AssignmentTypeTask,
			Title:       ft.Title,
			Status:      shpankids.StatusOpen,
			Description: ft.Description,
		})
	})

	assignmentsById := functional.SliceToMapNoErr(tasks, func(t *shpankids.Assignment) string {
		return t.Id
	})

	userTaskRepo, err := NewUserTaskStatusRepository(ctx, m.kvs, userId)
	if err != nil {
		return shpanstream.NewErrorStream[shpankids.Assignment](err)
	}

	err = userTaskRepo.GetAllForDate(ctx, forDate).Consume(
		ctx,
		func(dr *functional.Entry[string, dbUserTaskStatus]) {
			foundTask, found := assignmentsById[dr.Key]
			if found {
				foundTask.Status = dr.Value.Status
			}
		},
	)
	if err != nil {
		return shpanstream.NewErrorStream[shpankids.Assignment](err)
	}

	return shpanstream.Just[shpankids.Assignment](functional.MapSliceUnPtr(functional.MapValues(assignmentsById))...)
}

func (m *managerImpl) getUserTaskStatesForDateRange(
	ctx context.Context,
	from datekvs.Date,
	to datekvs.Date,
	familyTasks []shpankids.FamilyTaskDto,
	userId string,
) shpanstream.Stream[shpankids.TaskStats] {
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
		return shpanstream.NewErrorStream[shpankids.TaskStats](err)
	}

	return shpanstream.MapStreamWhileFilteringWithError[datekvs.Date, shpankids.TaskStats](
		datekvs.NewDateRangeStream(from, to),
		func(ctx context.Context, dt *datekvs.Date) (*shpankids.TaskStats, error) {
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

			ret := &shpankids.TaskStats{
				UserId:          userId,
				ForDate:         dt.Time,
				TotalTasksCount: len(userAssignableTasksByTaskId),
				DoneTasksCount:  0,
			}

			err := userTaskRepo.GetAllForDate(ctx, *dt).Consume(ctx, func(dr *functional.Entry[string, dbUserTaskStatus]) {

				if _, found := userAssignableTasksByTaskId[dr.Key]; found {
					if dr.Value.Status == shpankids.StatusDone {
						ret.DoneTasksCount++
					}
				}
			})
			if err != nil {
				return nil, err
			}
			return ret, nil
		})

}

func (m *managerImpl) UpdateTaskStatus(
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
