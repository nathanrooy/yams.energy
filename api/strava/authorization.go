package stravaAPI

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"yams/services/strava"
)

type StravaResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
	Athlete      struct {
		ID        int64  `json:"id"`
		FirstName string `json:"firstname"`
		LastName  string `json:"lastname"`
	} `json:"athlete"`
}

func getTokensFromCode(code string) StravaResponse {

	// get user tokens from strava
	params := url.Values{}
	params.Add("client_id", os.Getenv("STRAVA_CLIENT_ID"))
	params.Add("client_secret", os.Getenv("STRAVA_CLIENT_SECRET"))
	params.Add("code", code)
	params.Add("grant_type", "authorization_code")
	resp, err := http.PostForm("https://www.strava.com/oauth/token", params)
	if err != nil {
		log.Printf("error getting user tokens from strava")
	}
	defer resp.Body.Close()

	// parse response from strava
	body, err := io.ReadAll(resp.Body)
	log.Printf("> body: %v", string(body))
	if err != nil {
		// handle error
		log.Println("failed to parse strava response")
	}
	var stravaResponse StravaResponse
	err = json.Unmarshal(body, &stravaResponse)
	if err != nil {
		// handle error
		log.Println("unmarshal strava response error")
	}
	return stravaResponse
}

func Authorization(w http.ResponseWriter, r *http.Request) {
	log.Println("authorizing new strava user...")
	if r.Method == "GET" {
		if r.URL.Query().Get("error") == "access_denied" {
			log.Printf("access denied")
		}

		var code string = r.URL.Query().Get("code")
		if code == "" {
			log.Println("> failed to authenticate new user (no code from Strava)")
			http.Redirect(w, r, "/authorization-fail", http.StatusSeeOther)

		} else {
			stravaResponse := getTokensFromCode(code)
			strava.AddNewUser(
				stravaResponse.Athlete.ID,
				stravaResponse.AccessToken,
				stravaResponse.RefreshToken,
				stravaResponse.ExpiresAt,
			)
			log.Println("> authorization successful")
			http.Redirect(w, r, "/authorization-pass", http.StatusSeeOther)
		}
	}
}
