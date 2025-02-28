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

type CreateProblemAnswerDto struct {
	Title       string
	Description string
	Correct     bool
}

type ProblemAnswerDto struct {
	Id          string
	Title       string
	Description string
	Correct     bool
}

type ProblemDto struct {
	ProblemId   string
	Title       string
	Description string
	Created     time.Time
	Hints       []string
	Explanation string
	Answers     []ProblemAnswerDto
}

type CreateProblemDto struct {
	Title       string
	Description string
	Hints       []string
	Explanation string
	Answers     []CreateProblemAnswerDto
}

type ProblemSetDto struct {
	ProblemSetId string
	Title        string
	Description  string
	Created      time.Time
}

type ProblemSolutionDto struct {
	ProblemId        string
	Correct          bool
	SelectedAnswerId string
}

type CreateProblemSetDto struct {
	ProblemSetId string
	Title        string
	Description  string
}

type FamilyAssignmentStatus string

const (
	FamilyAssignmentStatusActive  FamilyAssignmentStatus = "active"
	FamilyAssignmentStatusDeleted FamilyAssignmentStatus = "deleted"
)

type FamilyManager interface {
	GetProblemSetManagerForUser(ctx context.Context, forUserId string) (ProblemSetManager, error)

	CreateFamily(ctx context.Context, familyId string, familyName string, memberUserIds []string, adminUserIds []string) error
	GetFamily(ctx context.Context, familyId string) (*FamilyDto, error)
	FindFamily(ctx context.Context, familyId string) (*FamilyDto, error)

	CreateFamilyTask(ctx context.Context, familyId string, familyTask FamilyTaskDto) error
	ListFamilyTasks(ctx context.Context, familyId string) shpanstream.Stream[FamilyTaskDto]
	DeleteFamilyTask(ctx context.Context, familyId string, familyTaskId string) error
}
