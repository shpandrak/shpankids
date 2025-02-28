package shpankids

import (
	"context"
	"shpankids/infra/database/datekvs"
	"shpankids/infra/shpanstream"
	"shpankids/openapi"
)

type ProblemSetManager interface {
	CreateProblemsInSet(
		ctx context.Context,
		problemSetId string,
		args []CreateProblemDto,
	) error

	CreateProblemSet(
		ctx context.Context,
		args CreateProblemSetDto,
	) error

	ListProblemSets(
		ctx context.Context,
	) shpanstream.Stream[ProblemSetDto]

	ListProblemsForProblemSet(
		ctx context.Context,
		problemSetId string,
		includingArchived bool,
	) shpanstream.Stream[ProblemDto]

	GenerateNewProblems(
		ctx context.Context,
		problemSetId string,
		additionalRequestText string,
	) shpanstream.Stream[openapi.ApiProblemForEdit]

	RefineProblems(
		ctx context.Context,
		problemSetId string,
		origProblems shpanstream.Stream[openapi.ApiProblemForEdit],
		refineInstructions string,
	) shpanstream.Stream[openapi.ApiProblemForEdit]

	SubmitProblemAnswer(
		ctx context.Context,
		problemSetId string,
		problemId string,
		forDate datekvs.Date,
		answerId string,
	) (bool, string, *ProblemDto, error)

	ListProblemSetSolutionsForDate(
		ctx context.Context,
		problemSetId string,
		forDate datekvs.Date,
	) shpanstream.Stream[ProblemSolutionDto]

	ListProblemsSolutions(
		ctx context.Context,
		problemSetId string,
	) shpanstream.Stream[openapi.ApiUserProblemSolution]

	GetProblem(
		ctx context.Context,
		problemSetId string,
		problemId string,
	) (*ProblemDto, error)
}
