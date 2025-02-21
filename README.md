
# Running locally

To run the application locally, you need to create secrets.json file in the root of the project with the following content:

```json
{
  "oAuth": {
    "clientId": "[googleOAuthClientId]",
    "clientSecret": "[googleOAuthClientSecret]"
  },
  "gemini": {
    "apiKey": "[googleGeminiApiKey]"
  }
}
```

in order to deploy the application to gcp, and update the secret, run:

```bash
gcloud secrets versions add SHPANSECRET --data-file=secrets.json
```