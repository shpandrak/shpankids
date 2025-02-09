package google

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

const clientId = "[clientId]"
const clientSecret = "[clientSecret]"
const oAuthStateStringGl = "whatever"

var (
	oauthConfGl = &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
)

// Helper to determine the request scheme (HTTP or HTTPS)
func getScheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

func handleLogin(w http.ResponseWriter, r *http.Request, oauthConf *oauth2.Config, oauthStateString string) {

	oauthConfGl.RedirectURL = fmt.Sprintf("%s://%s/callback-gl", getScheme(r), r.Host)

	theAuthURL, err := url.Parse(oauthConf.Endpoint.AuthURL)
	if err != nil {
		slog.Error(fmt.Sprintf("Parse: %s", err.Error()))
	} else {

		slog.Info(theAuthURL.String())
		parameters := url.Values{}
		parameters.Add("client_id", oauthConf.ClientID)
		parameters.Add("scope", strings.Join(oauthConf.Scopes, " "))
		parameters.Add("redirect_uri", oauthConf.RedirectURL)
		parameters.Add("response_type", "code")
		parameters.Add("state", oauthStateString)
		theAuthURL.RawQuery = parameters.Encode()
		urlToRedirect := theAuthURL.String()
		slog.Info(urlToRedirect)
		http.Redirect(w, r, urlToRedirect, http.StatusTemporaryRedirect)
	}
}

var authCallbackFunc func(r *http.Request, w http.ResponseWriter, email string)

func RegisterCallbacks(router *mux.Router, authCallback func(r *http.Request, w http.ResponseWriter, email string)) {
	authCallbackFunc = authCallback
	router.HandleFunc("/login-gl", handleGoogleLogin)
	router.HandleFunc("/callback-gl", callBackFromGoogle)

}

func handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	handleLogin(w, r, oauthConfGl, oAuthStateStringGl)
}

/*
callBackFromGoogle Function
*/
func callBackFromGoogle(w http.ResponseWriter, r *http.Request) {
	slog.Info("Callback-gl..")

	state := r.FormValue("state")
	slog.Info(state)

	code := r.FormValue("code")
	slog.Info(code)

	if code == "" {
		slog.Warn("Code not found..")
		w.Write([]byte("Code Not Found to provide AccessToken..\n"))
		reason := r.FormValue("error_reason")
		if reason == "user_denied" {
			w.Write([]byte("User has denied Permission.."))
		}
		// User has denied access..
		// http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	} else {
		token, err := oauthConfGl.Exchange(oauth2.NoContext, code)
		if err != nil {
			slog.Error("oauthConfGl.Exchange() failed with " + err.Error() + "\n")
			return
		}
		slog.Info("TOKEN>> AccessToken>> " + token.AccessToken)
		slog.Info("TOKEN>> Expiration Time>> " + token.Expiry.String())
		slog.Info("TOKEN>> RefreshToken>> " + token.RefreshToken)

		resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + url.QueryEscape(token.AccessToken))
		if err != nil {
			slog.Error("Get: " + err.Error() + "\n")
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		defer resp.Body.Close()

		response, err := io.ReadAll(resp.Body)
		if err != nil {
			slog.Error("ReadAll: " + err.Error() + "\n")
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		var parsedResp *googleAuthRespDto
		err = json.Unmarshal(response, &parsedResp)
		if err != nil {
			slog.Error("Unmarshal: " + err.Error() + "\n")
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		slog.Info("parseResponseBody: " + string(response) + "\n")

		authCallbackFunc(r, w, parsedResp.Email)
		//w.Write([]byte("Hello, I'm protected\n"))
		//w.Write(response)
		return
	}
}

type googleAuthRespDto struct {
	Id            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Picture       string `json:"picture"`
}
