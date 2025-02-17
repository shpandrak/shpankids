package shpankids

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
)

type OAuthSecret struct {
	ClientId     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

type GeminiSecret struct {
	ApiKey string `json:"apiKey"`
}

type Secrets struct {
	OAuth  OAuthSecret  `json:"oAuth"`
	Gemini GeminiSecret `json:"gemini"`
}

var secrets Secrets

func DetectSecrets() error {

	var shpanSecretsRaw []byte
	// if file secrets.json exists, read it and override the env var
	if _, err := os.Stat("secrets.json"); err == nil {
		shpanSecretsRaw, err = os.ReadFile("secrets.json")
		if err != nil {
			return fmt.Errorf("failed to open secrets.json: %v", err)
		}
		slog.Info("shpankids secrets loaded from file")
	} else {
		shpanSecretsStr := os.Getenv("SHPAN_SECRETS")
		if shpanSecretsStr == "" {
			return fmt.Errorf("SHPAN_SECRETS evn var not found")
		}
		shpanSecretsRaw = []byte(shpanSecretsStr)
		slog.Info("shpankids secrets loaded from env var")
	}

	var theSecrets Secrets
	err := json.Unmarshal(shpanSecretsRaw, &theSecrets)
	if err != nil {
		return fmt.Errorf("failed to unmarshal SHPAN_SECRETS: %v", err)
	}
	secrets = theSecrets
	return nil

}

func GetSecrets() Secrets {
	return secrets
}
