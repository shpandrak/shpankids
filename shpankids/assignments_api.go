package shpankids

import (
	"context"
	"shpankids/infra/database/datekvs"
	"shpankids/infra/shpanstream"
	"time"
)

type AssignmentStatus string

const (
	StatusOpen       AssignmentStatus = "open"
	StatusDone       AssignmentStatus = "done"
	StatusBlocked    AssignmentStatus = "blocked"
	StatusIrrelevant AssignmentStatus = "irrelevant"
)

type AssignmentType string

const (
	AssignmentTypeTask       AssignmentType = "task"
	AssignmentTypeProblemSet AssignmentType = "problemSet"
)

type DailyAssignmentDto struct {
	Id          string
	ForDate     datekvs.Date
	Type        AssignmentType
	Title       string
	Status      AssignmentStatus
	Description string
}

type CreateAssignmentArgsDto struct {
	Id            string
	Type          AssignmentType
	Title         string
	NumberOfParts int
	Description   string
}

type AssignmentStats struct {
	UserId          string
	ForDate         time.Time
	TotalTasksCount int
	DoneTasksCount  int
}

type AssignmentManager interface {
	ArchiveUserAssignment(ctx context.Context, forUserId string, assignmentId string) error
	CreateNewAssignment(ctx context.Context, forUserId string, args CreateAssignmentArgsDto) error

	ReportTaskProgress(ctx context.Context, forUserId string, forDate datekvs.Date, assignmentId string, partsDelta int, comment string) error

	ListMyAssignmentsForToday(ctx context.Context) shpanstream.Stream[DailyAssignmentDto]
	GetAssignmentStats(ctx context.Context, fromDate datekvs.Date, toDate datekvs.Date) shpanstream.Stream[AssignmentStats]
	UpdateAssignmentStatus(ctx context.Context, forDay time.Time, taskId string, status AssignmentStatus, comment string) error
}
