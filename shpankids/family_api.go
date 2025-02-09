package shpankids

import (
	"context"
)

type FamilyDto struct {
	Id   string
	Name string
}

type FamilyTaskDto struct {
	TaskId      string
	Title       string
	Description string
	MemberIds   []string
}

type FamilyManager interface {
	CreateFamily(ctx context.Context, familyId string, familyName string, memberUserIds []string) error
	CreateFamilyTask(ctx context.Context, familyId string, familyTask FamilyTaskDto) error
	FindFamily(ctx context.Context, familyId string) (*FamilyDto, error)
	FamilyTasks(ctx context.Context, familyId string) ([]FamilyTaskDto, error)
}
