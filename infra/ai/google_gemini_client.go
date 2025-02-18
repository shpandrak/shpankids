package ai

import (
	"context"
	"fmt"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func GetClient(ctx context.Context, apiKey string) error {

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash")
	model.SystemInstruction = genai.NewUserContent(genai.Text("You are a cat. Your name is Neko."))
	resp, err := model.GenerateContent(ctx, genai.Text("Good morning! How are you?"))
	if err != nil {
		return err
	}
	printResponse(resp)
	return nil
}

func printResponse(resp *genai.GenerateContentResponse) {
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				fmt.Println(part)
			}
		}
	}
	fmt.Println("---")
}
