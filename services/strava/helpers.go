package strava

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"yams/app"
	"yams/database"
)

type Tokens struct {
	AccessToken  string `json:"access_token"`
	AthleteId    int64
	ExpiresAt    int64  `json:"expires_at"`
	RefreshToken string `json:"refresh_token"`
}

type Activity struct {
	Calories    float64    `json:"calories"`
	Description string     `json:"description"`
	Id          int64      `json:"id"`
	Manual      bool       `json:"manual" default:"false"`
	StartDate   string     `json:"start_date"`
	StartLatLng [2]float64 `json:"start_latlng"`
	Trainer     bool       `json:"trainer" default:"true"`
	Type        string     `json:"type"`
}

func AddNewUser(athleteId int64, accessToken string, refreshToken string, expiresAt int64) {
	log.Printf("adding new strava user: %v\n", athleteId)
	db := database.Connect()
	defer db.Close()

	tokens := Tokens{
		AccessToken:  accessToken,
		AthleteId:    athleteId,
		ExpiresAt:    expiresAt,
		RefreshToken: refreshToken,
	}

	updateUserTokens(db, tokens)
	database.AddEvent(db, map[string]string{"event_type": "user_subscribed", "platform": "strava", "id": strconv.FormatInt(athleteId, 10)})
}

func updateUserTokens(db *sql.DB, tokens Tokens) {
	sql := `INSERT INTO %s.subscribers.strava (id, access_token, refresh_token, expires_at) VALUES($1, $2, $3, $4) ON CONFLICT (id) DO UPDATE SET access_token = $2, refresh_token = $3, expires_at = $4;`
	stmt, err := db.Prepare(fmt.Sprintf(sql, os.Getenv("DB_DATABASE")))
	if err != nil {
		log.Printf("> failed to prepare insert: %v", err)
	}
	result, err := stmt.Exec(tokens.AthleteId, tokens.AccessToken, tokens.RefreshToken, tokens.ExpiresAt)
	if err != nil {
		log.Printf("> db-err: %s\n", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 1 {
		log.Printf("> successfully updated tokens for strava user \"%v\"\n", tokens.AthleteId)
	} else {
		log.Printf("> failed to update tokens for strava user \"%v\"\n", tokens.AthleteId)
	}
}

func DeleteUser(athleteId int64) {
	log.Printf("> Deleting strava user: %d\n", athleteId)

	db := database.Connect()
	defer db.Close()

	sql := `DELETE FROM %s.subscribers.strava WHERE id = $1`
	stmt, err := db.Prepare(fmt.Sprintf(sql, os.Getenv("DB_DATABASE")))
	if err != nil {
		log.Printf("> failed to prepare delete query: %v", err)
	}
	result, err := stmt.Exec(athleteId)
	if err != nil {
		log.Printf("db-err: %s\n", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 1 {
		log.Printf("successfully removed user \"%v\" from \"strava.subscribers\"\n", athleteId)
	} else {
		log.Printf("failed to remove user \"%v\" from \"strava.subscribers\"\n", athleteId)
	}

	database.AddEvent(db, map[string]string{"event_type": "user_unsubscribed", "platform": "strava", "id": strconv.FormatInt(athleteId, 10)})
}

func refreshUserTokens(tokens Tokens) Tokens {

	if tokens.ExpiresAt > time.Now().Unix() {
		log.Printf("> current user tokens are still valid\n")
		return tokens
	} else {
		log.Printf("> tokens have expired. contacting strava for latest user tokens\n")

		// refresh tokens
		params := url.Values{}
		params.Add("client_id", os.Getenv("STRAVA_CLIENT_ID"))
		params.Add("client_secret", os.Getenv("STRAVA_CLIENT_SECRET"))
		params.Add("refresh_token", tokens.RefreshToken)
		params.Add("grant_type", "refresh_token")
		resp, err := http.PostForm("https://www.strava.com/oauth/token", params)
		if err != nil {
			log.Printf("error refreshing user tokens from strava")
		}
		defer resp.Body.Close()

		// parse response from strava
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			// handle error
		}
		var refreshedTokens Tokens
		err = json.Unmarshal(body, &refreshedTokens)
		if err != nil {
			log.Printf("> token unmarshal error: %v", err)
		}
		refreshedTokens.AthleteId = tokens.AthleteId
		if refreshedTokens.ExpiresAt > tokens.ExpiresAt {
			log.Printf("> tokens have been refreshed\n")
		}
		return refreshedTokens
	}
}

func getUserTokens(db *sql.DB, athleteId int64) Tokens {
	log.Printf("> getting user tokens\n")
	sql := `SELECT id, access_token, refresh_token, expires_at FROM %s.subscribers.strava WHERE id = $1 LIMIT 1`
	stmt, _ := db.Prepare(fmt.Sprintf(sql, os.Getenv("DB_DATABASE")))

	var tokens Tokens
	if err := stmt.QueryRow(athleteId).Scan(&tokens.AthleteId, &tokens.AccessToken, &tokens.RefreshToken, &tokens.ExpiresAt); err != nil {
		log.Printf("> error getting user tokens: %v\n", err)
	}
	return tokens
}

func getActivity(activityId int64, tokens Tokens) Activity {

	log.Print("getting user activity\n")

	// Create a new HTTP request
	req, err := http.NewRequest("GET", fmt.Sprintf("https://www.strava.com/api/v3/activities/%v", activityId), nil)
	if err != nil {
		panic(err)
	}

	// Set the Authorization header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", tokens.AccessToken))

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Read the response body.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	// Munge response body.
	var activity Activity
	_ = json.Unmarshal(body, &activity)
	return activity
}

func modifyActivity(activity Activity, tokens Tokens) {
	log.Printf("> modifying activity: %v for user: %v\n", activity.Id, tokens.AthleteId)

	// create payload
	payload := map[string]string{
		"description": activity.Description,
	}

	// set the payload
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	// Create a new HTTP request
	url := fmt.Sprintf("https://www.strava.com/api/v3/activities/%v", activity.Id)
	req, err := http.NewRequest("PUT", url, bytes.NewReader(jsonPayload))
	if err != nil {
		panic(err)
	}

	// Set the Authorization header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", tokens.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("> error sending put request to strava: %v", err)
	}

	// Check the response status code.
	log.Printf("> response code: %v\n", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		log.Printf("unexpected response status code: %d", resp.StatusCode)
	}
}

func ProcessNewActivity(athleteId int64, activityId int64) {
	log.Printf("> adding calorie description for strava user: %v, activity = %v", athleteId, activityId)
	tStartMs := time.Now().UnixMilli()

	db := database.Connect()
	defer db.Close()

	oldTokens := getUserTokens(db, athleteId)
	tokens := refreshUserTokens(oldTokens)

	// persist latest user tokens (if necessary)
	if tokens.ExpiresAt > oldTokens.ExpiresAt {
		log.Printf("> persisting updated tokens\n")
		updateUserTokens(db, tokens)
	}

	// Get the actual activity object from Strava
	activity := getActivity(activityId, tokens)

	// Only add a calorie description for certain activities
	if activity.Manual == true || activity.Trainer == true || activity.Type == "VirtualRide" {
		log.Printf("Activity %v was manually created or indoor. Skipping...\n", activityId)
	} else if strings.Contains(activity.Description, "yams") {
		log.Printf("Activity %v already has weather information\n", activityId)
	} else if activity.StartLatLng == [2]float64{} {
		log.Printf("No position present for activity %v\n", activityId)
	} else if activity.Calories <= 25 {
		log.Printf("Less than 25 calories for activity %v\n", activityId)
	} else {

		calorieDescription := app.GenerateDescription(activity.Calories)
		if activity.Description != "" {
			activity.Description = strings.TrimRight(activity.Description, " ") + "\n" + calorieDescription
		} else {
			activity.Description = calorieDescription
		}

		// Update Strava activity
		modifyActivity(activity, tokens)
		tDelta := time.Now().UnixMilli() - tStartMs

		log.Printf("> finished adding description for %v, activity: %v in %v ms\n", athleteId, activityId, tDelta)
	}

}
