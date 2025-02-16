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
	Status      FamilyAssignmentStatus
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
	Hints        []string
	Explanation  string
	Alternatives []ProblemAlternativeDto
}

type FamilyProblemSetDto struct {
	ProblemSetId string
	Title        string
	Description  string
	Created      time.Time
	Status       FamilyAssignmentStatus
	StatusDate   time.Time
}

type FamilyAssignmentStatus string

const (
	FamilyAssignmentStatusActive  FamilyAssignmentStatus = "active"
	FamilyAssignmentStatusDeleted FamilyAssignmentStatus = "deleted"
)

type FamilyManager interface {
	CreateFamily(ctx context.Context, familyId string, familyName string, memberUserIds []string, adminUserIds []string) error
	GetFamily(ctx context.Context, familyId string) (*FamilyDto, error)
	FindFamily(ctx context.Context, familyId string) (*FamilyDto, error)

	CreateFamilyTask(ctx context.Context, familyId string, familyTask FamilyTaskDto) error
	ListFamilyTasks(ctx context.Context, familyId string) shpanstream.Stream[FamilyTaskDto]
	DeleteFamilyTask(ctx context.Context, familyId string, familyTaskId string) error

	CreateFamilyProblem(ctx context.Context, familyId string, forUserId string, problemSetId string, familyProblem FamilyProblemDto) error
	CreateFamilyProblemSet(ctx context.Context, familyId string, forUserId string, familyProblemSet FamilyProblemSetDto) error
	ListFamilyProblemSetsForUser(ctx context.Context, familyId string, userId string) shpanstream.Stream[FamilyProblemSetDto]
	ListFamilyProblemsForUser(ctx context.Context, familyId string, userId string, problemSetId string) shpanstream.Stream[FamilyProblemDto]
}
