package shpankids

import (
	"context"
	"shpankids/infra/database/datekvs"
	"shpankids/infra/shpanstream"
	"time"
)

type TaskStatus string

const (
	StatusOpen       TaskStatus = "open"
	StatusDone       TaskStatus = "done"
	StatusBlocked    TaskStatus = "blocked"
	StatusIrrelevant TaskStatus = "irrelevant"
)

type Task struct {
	Id          string
	Title       string
	Description string
	DueDate     time.Time
	Status      TaskStatus
}

type TaskStats struct {
	UserId          string
	ForDate         time.Time
	TotalTasksCount int
	DoneTasksCount  int
}

type TaskManager interface {
	GetTasksForDate(ctx context.Context, date time.Time) ([]Task, error)
	GetTaskStats(ctx context.Context, fromDate datekvs.Date, toDate datekvs.Date) shpanstream.Stream[TaskStats]
	UpdateTaskStatus(ctx context.Context, forDay time.Time, taskId string, status TaskStatus, comment string) error
}
