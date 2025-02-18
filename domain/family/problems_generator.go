package family

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"shpankids/infra/shpanstream"
	"shpankids/infra/util/functional"
	"shpankids/openapi"
	"shpankids/shpankids"
	"strings"
)

func generateProblems(
	ctx context.Context,
	forUserId string,
	problemSet shpankids.FamilyProblemSetDto,
	examples shpanstream.Stream[shpankids.FamilyProblemDto],
	additionalRequestText string,
) shpanstream.Stream[openapi.ApiProblemForEdit] {
	client, err := genai.NewClient(ctx, option.WithAPIKey(shpankids.GetSecrets().Gemini.ApiKey))
	if err != nil {
		return shpanstream.NewErrorStream[openapi.ApiProblemForEdit](err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.0-flash")
	//model.SetTemperature(1)
	//model.SetTopK(40)
	//model.SetTopP(0.95)
	//model.SetMaxOutputTokens(8192)
	model.ResponseMIMEType = "application/json"
	model.ResponseSchema = &genai.Schema{
		Type:        genai.TypeArray,
		Description: "List of next problems to challenge the family member with on the topic",
		Items: &genai.Schema{
			Type: genai.TypeObject,
			Required: []string{
				"title",
				"answers",
			},
			Properties: map[string]*genai.Schema{
				"title": {
					Type:        genai.TypeString,
					Description: "problem title and question",
					Nullable:    false,
				},
				"description": {
					Type:        genai.TypeString,
					Description: "problem description",
					Nullable:    true,
				},
				"answers": {
					Type:     genai.TypeArray,
					Nullable: false,
					Items: &genai.Schema{
						Type: genai.TypeObject,
						Required: []string{
							"title",
							"isCorrect",
						},
						Properties: map[string]*genai.Schema{
							"title": {
								Type:        genai.TypeString,
								Description: "answer title",
								Nullable:    false,
							},
							"description": {
								Type:        genai.TypeString,
								Description: "answer description",
								Nullable:    true,
							},
							"isCorrect": {
								Type:        genai.TypeBoolean,
								Description: "is this answer correct",
								Nullable:    false,
							},
						},
					},
				},
			},
		},
	}

	strs, err := shpanstream.MapStreamWithError(
		examples.Limit(5),
		func(ctx context.Context, dto *shpankids.FamilyProblemDto) (*string, error) {
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

	sampleFullJson := fmt.Sprintf("[%s]", strings.Join(strs, ","))

	if additionalRequestText != "" {
		additionalRequestText = fmt.Sprintf(". Additional request:%s", additionalRequestText)
	}
	session := model.StartChat()
	session.History = []*genai.Content{
		{
			Role: "user",
			Parts: []genai.Part{
				genai.Text(fmt.Sprintf(
					"Generate a list of problems to challenge the family member %s, "+
						"on the topic of %s %s. Make the outputs in JSON format.",
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
			"base on the examples provided, please suggest next problems for family member %s, "+
				"on the same topic of %s. Make the outputs in JSON format. "+
				"Each problem should have a title and a list of answers, "+
				"from which only one is correct.%s",
			forUserId,
			problemSet.Title,
			additionalRequestText,
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
