package shpankids

import (
	"context"
	"shpankids/infra/database/datekvs"
	"shpankids/infra/shpanstream"
	"shpankids/openapi"
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

type FamilyProblemDto struct {
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

type FamilyProblemSetDto struct {
	ProblemSetId string
	Title        string
	Description  string
	Created      time.Time
	Status       FamilyAssignmentStatus
	StatusDate   time.Time
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
	CreateFamily(ctx context.Context, familyId string, familyName string, memberUserIds []string, adminUserIds []string) error
	GetFamily(ctx context.Context, familyId string) (*FamilyDto, error)
	FindFamily(ctx context.Context, familyId string) (*FamilyDto, error)

	CreateFamilyTask(ctx context.Context, familyId string, familyTask FamilyTaskDto) error
	ListFamilyTasks(ctx context.Context, familyId string) shpanstream.Stream[FamilyTaskDto]
	DeleteFamilyTask(ctx context.Context, familyId string, familyTaskId string) error

	CreateProblemsInSet(ctx context.Context, familyId string, forUserId string, problemSetId string, familyProblem []CreateProblemDto) error
	CreateProblemSet(ctx context.Context, familyId string, forUserId string, familyProblemSet CreateProblemSetDto) error
	ListProblemSetsForUser(ctx context.Context, familyId string, userId string) shpanstream.Stream[FamilyProblemSetDto]

	ListProblemsForProblemSet(
		ctx context.Context,
		familyId string,
		userId string,
		problemSetId string,
		includingArchived bool,
	) shpanstream.Stream[FamilyProblemDto]

	GenerateNewProblems(
		ctx context.Context,
		familyId string,
		userId string,
		problemSetId string,
		additionalRequestText string,
	) shpanstream.Stream[openapi.ApiProblemForEdit]

	RefineProblems(
		ctx context.Context,
		familyId string,
		userId string,
		problemSetId string,
		origProblems shpanstream.Stream[openapi.ApiProblemForEdit],
		refineInstructions string,
	) shpanstream.Stream[openapi.ApiProblemForEdit]

	SubmitProblemAnswer(
		ctx context.Context,
		familyId string,
		userId string,
		problemSetId string,
		problemId string,
		forDate datekvs.Date,
		answerId string,
	) (bool, string, *FamilyProblemDto, error)

	ListProblemSetSolutionsForDate(
		ctx context.Context,
		familyId string,
		userId string,
		problemSetId string,
		forDate datekvs.Date,
	) shpanstream.Stream[ProblemSolutionDto]

	ListUserProblemsSolutions(
		ctx context.Context,
		familyId string,
		problemSetId string,
		userId string,
	) shpanstream.Stream[openapi.ApiUserProblemSolution]

	GetProblem(
		ctx context.Context,
		familyId string,
		userId string,
		problemSetId string,
		problemId string,
	) (*FamilyProblemDto, error)
}
