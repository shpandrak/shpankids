package shpankids

import (
	"context"
	"time"
)

type Status string

const (
	StatusOpen       Status = "open"
	StatusDone       Status = "done"
	StatusBlocked    Status = "blocked"
	StatusIrrelevant Status = "irrelevant"
)

type Task struct {
	Id          string
	Title       string
	Description string
	DueDate     time.Time
	Status      Status
}

type TaskStats struct {
	UserId          string
	ForDate         time.Time
	TotalTasksCount int
	DoneTasksCount  int
}

type Manager interface {
	GetTasksForDate(ctx context.Context, date time.Time) ([]Task, error)
	GetTaskStats(ctx context.Context, fromDate time.Time, toDate time.Time) ([]TaskStats, error)
	UpdateTaskStatus(ctx context.Context, forDay time.Time, taskId string, status Status, comment string) error
}
