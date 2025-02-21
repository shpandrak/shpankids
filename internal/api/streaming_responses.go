package api

import (
	"context"
	"net/http"
	"shpankids/infra/shpanstream"
	"shpankids/openapi"
)

type streamingAssignments struct {
	stream shpanstream.Stream[openapi.ApiAssignment]
	ctx    context.Context
}

func (s *streamingAssignments) VisitListAssignmentsResponse(w http.ResponseWriter) error {
	return shpanstream.StreamToJsonResponseWriter(s.ctx, w, s.stream)
}

type streamingGetStatsResponseObject struct {
	stream shpanstream.Stream[openapi.ApiTaskStats]
	ctx    context.Context
}

func (s *streamingGetStatsResponseObject) VisitGetStatsResponse(w http.ResponseWriter) error {
	return shpanstream.StreamToJsonResponseWriter(s.ctx, w, s.stream)
}

type streamingProblemSets struct {
	stream shpanstream.Stream[openapi.ApiProblemSet]
	ctx    context.Context
}

func (s *streamingProblemSets) VisitListUserFamilyProblemSetsResponse(w http.ResponseWriter) error {
	return shpanstream.StreamToJsonResponseWriter(s.ctx, w, s.stream)
}

type streamingProblemsForEdit struct {
	stream shpanstream.Stream[openapi.ApiProblemForEdit]
	ctx    context.Context
}

func (s *streamingProblemsForEdit) VisitRefineProblemsResponse(w http.ResponseWriter) error {
	return shpanstream.StreamToJsonResponseWriter(s.ctx, w, s.stream)
}

func (s *streamingProblemsForEdit) VisitListProblemSetProblemsResponse(w http.ResponseWriter) error {
	return shpanstream.StreamToJsonResponseWriter(s.ctx, w, s.stream)
}
func (s *streamingProblemsForEdit) VisitGenerateProblemsResponse(w http.ResponseWriter) error {
	return shpanstream.StreamToJsonResponseWriter(s.ctx, w, s.stream)
}
