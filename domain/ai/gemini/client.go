package gemini

import (
	"context"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"shpankids/shpankids"
	"sync"
)

var client *genai.Client
var once sync.Once

func GetClient(ctx context.Context) (*genai.Client, error) {
	var innerErr error
	once.Do(func() {
		client, innerErr = genai.NewClient(ctx, option.WithAPIKey(shpankids.GetSecrets().Gemini.ApiKey))
	})
	if innerErr != nil {
		return nil, innerErr
	}
	return client, nil
}

func GetDefaultModel(ctx context.Context) (*genai.GenerativeModel, error) {
	c, err := GetClient(ctx)
	if err != nil {
		return nil, err
	}
	model := c.GenerativeModel("gemini-2.0-flash")
	//model.SetTemperature(1)
	//model.SetTopK(40)
	//model.SetTopP(0.95)
	//model.SetMaxOutputTokens(8192)
	return model, err

}
