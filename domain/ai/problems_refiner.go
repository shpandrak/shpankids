package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/generative-ai-go/genai"
	"shpankids/domain/ai/gemini"
	"shpankids/infra/shpanstream"
	"shpankids/infra/util/functional"
	"shpankids/openapi"
	"shpankids/shpankids"
	"strings"
)

func RefineProblems(
	ctx context.Context,
	forUserId string,
	problemSet shpankids.FamilyProblemSetDto,
	origProblems shpanstream.Stream[openapi.ApiProblemForEdit],
	refineInstructions string,
) shpanstream.Stream[openapi.ApiProblemForEdit] {
	model, err := gemini.GetDefaultModel(ctx)
	if err != nil {
		return shpanstream.NewErrorStream[openapi.ApiProblemForEdit](err)
	}
	model.ResponseMIMEType = "application/json"
	model.ResponseSchema = apiProblemsForEditArrSchema()

	model.ResponseMIMEType = "application/json"
	model.ResponseSchema = apiProblemsForEditArrSchema()

	strs, err := shpanstream.MapStreamWithError(
		origProblems,
		func(ctx context.Context, dto *openapi.ApiProblemForEdit) (*string, error) {
			marshal, err := json.Marshal(dto)
			if err != nil {
				return nil, err
			}
			return functional.ValueToPointer(string(marshal)), nil
		},
	).CollectFilterNil(ctx)
	if err != nil {
		return shpanstream.NewErrorStream[openapi.ApiProblemForEdit](err)
	}

	if len(strs) == 0 {
		return shpanstream.EmptyStream[openapi.ApiProblemForEdit]()
	}

	sampleFullJson := fmt.Sprintf("[%s]", strings.Join(strs, ","))

	session := model.StartChat()
	session.History = []*genai.Content{
		{
			Role: "user",
			Parts: []genai.Part{
				genai.Text(fmt.Sprintf(
					"Generate a list of problems to challenge the family member %s, "+
						"on the topic of %s %s. Make the outputs in JSON format."+
						"Each problem should have a title and a list of answers, "+
						"from which only one is correct",
					forUserId,
					problemSet.Title,
					problemSet.Description,
				)),
			},
		},
		{
			Role: "model",
			Parts: []genai.Part{
				genai.Text(sampleFullJson),
			},
		},
	}

	resp, err := session.SendMessage(
		ctx,
		genai.Text(fmt.Sprintf(
			"please refine the problems you generated for family member %s, "+
				"on the same topic of %s. Make the outputs in JSON format. "+
				"please refine and return the problems according to the following request: %s",
			forUserId,
			problemSet.Title,
			refineInstructions,
		)),
	)

	if err != nil {
		return shpanstream.NewErrorStream[openapi.ApiProblemForEdit](err)
	}

	strFullRespJson := ""
	for _, part := range resp.Candidates[0].Content.Parts {
		strFullRespJson += fmt.Sprintf("%s", part)
	}
	var parsedJsonProblems []openapi.ApiProblemForEdit
	err = json.Unmarshal([]byte(strFullRespJson), &parsedJsonProblems)
	if err != nil {
		return shpanstream.NewErrorStream[openapi.ApiProblemForEdit](err)
	}
	return shpanstream.Just(parsedJsonProblems...)
}
