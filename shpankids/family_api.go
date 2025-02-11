package shpankids

import (
	"context"
	"time"
)

type FamilyDto struct {
	Id         string
	Name       string
	OwnerEmail string
	CreatedOn  time.Time
	Members    []FamilyMemberDto
}

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

type FamilyMemberDto struct {
	UserId string
	Role   Role
}

type FamilyTaskDto struct {
	TaskId      string
	Title       string
	Description string
	MemberIds   []string
	Status      FamilyTaskStatus
	StatusDate  time.Time
}

type FamilyTaskStatus string

const (
	FamilyTaskStatusActive  FamilyTaskStatus = "active"
	FamilyTaskStatusDeleted FamilyTaskStatus = "deleted"
)

type FamilyManager interface {
	CreateFamily(ctx context.Context, familyId string, familyName string, memberUserIds []string, adminUserIds []string) error
	CreateFamilyTask(ctx context.Context, familyId string, familyTask FamilyTaskDto) error
	FindFamily(ctx context.Context, familyId string) (*FamilyDto, error)
	ListFamilyTasks(ctx context.Context, familyId string) ([]FamilyTaskDto, error)
	DeleteFamilyTask(ctx context.Context, familyId string, familyTaskId string) error
	GetFamily(ctx context.Context, familyId string) (*FamilyDto, error)
}
