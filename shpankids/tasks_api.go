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
	AssignmentTypeTask    AssignmentType = "task"
	AssignmentTypeProblem AssignmentType = "problem"
)

type Assignment struct {
	Id          string
	ForDate     datekvs.Date
	Type        AssignmentType
	Title       string
	Status      AssignmentStatus
	Description string
}

type TaskStats struct {
	UserId          string
	ForDate         time.Time
	TotalTasksCount int
	DoneTasksCount  int
}

type AssignmentManager interface {
	ListAssignmentsForToday(ctx context.Context) ([]Assignment, error)
	ListAssignmentsForDate(ctx context.Context, date datekvs.Date) ([]Assignment, error)
	GetTaskStats(ctx context.Context, fromDate datekvs.Date, toDate datekvs.Date) shpanstream.Stream[TaskStats]
	UpdateTaskStatus(ctx context.Context, forDay time.Time, taskId string, status AssignmentStatus, comment string) error
}
