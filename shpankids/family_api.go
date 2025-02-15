package shpankids

import (
	"context"
	"shpankids/infra/shpanstream"
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
	Created     time.Time
	StatusDate  time.Time
}

type ProblemAlternativeDto struct {
	Title       string
	Description string
	Correct     bool
}

type FamilyProblemDto struct {
	ProblemId    string
	Title        string
	Description  string
	Created      time.Time
	Status       FamilyTaskStatus
	StatusDate   time.Time
	Hints        []string
	Explanation  string
	Alternatives []ProblemAlternativeDto
}

type FamilyTaskStatus string

const (
	FamilyTaskStatusActive  FamilyTaskStatus = "active"
	FamilyTaskStatusDeleted FamilyTaskStatus = "deleted"
)

type FamilyManager interface {
	CreateFamily(ctx context.Context, familyId string, familyName string, memberUserIds []string, adminUserIds []string) error
	CreateFamilyTask(ctx context.Context, familyId string, familyTask FamilyTaskDto) error
	CreateFamilyProblem(ctx context.Context, familyId string, forUserId string, familyProblem FamilyProblemDto) error
	FindFamily(ctx context.Context, familyId string) (*FamilyDto, error)
	ListFamilyTasks(ctx context.Context, familyId string) shpanstream.Stream[FamilyTaskDto]
	DeleteFamilyTask(ctx context.Context, familyId string, familyTaskId string) error
	GetFamily(ctx context.Context, familyId string) (*FamilyDto, error)
}
