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
}

type FamilyManager interface {
	CreateFamily(ctx context.Context, familyId string, familyName string, memberUserIds []string) error
	CreateFamilyTask(ctx context.Context, familyId string, familyTask FamilyTaskDto) error
	FindFamily(ctx context.Context, familyId string) (*FamilyDto, error)
	ListFamilyTasks(ctx context.Context, familyId string) ([]FamilyTaskDto, error)
}
