package stravaAPI

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"yams/services/strava"
)

type StravaPost struct {
	AspectType     string `json:"aspect_type"`
	EventTime      int64  `json:"event_time"`
	ObjectId       int64  `json:"object_id"`
	ObjectType     string `json:"object_type"`
	OwnerId        int64  `json:"owner_id"`
	SubscriptionId int64  `json:"subscription_id"`
	Updates        struct {
		Authorized string `json:"authorized"`
	} `json:"updates"`
}

func Webhook(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		processWebhookPOST(w, r)
	case "GET":
		processWebhookGET(w, r)
	}
}

func processWebhookPOST(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("read error: %v\n", err)
	}

	var stravaPost StravaPost
	err = json.Unmarshal(body, &stravaPost)
	if err != nil {
		log.Printf("> json error: %v\n", err)
	}

	w.WriteHeader(http.StatusOK)

	if stravaPost.AspectType == "create" && stravaPost.ObjectType == "activity" {
		strava.ProcessNewActivity(stravaPost.OwnerId, stravaPost.ObjectId)

	} else if stravaPost.Updates.Authorized == "false" {
		strava.DeleteUser(stravaPost.OwnerId)
	}
}

func processWebhookGET(w http.ResponseWriter, r *http.Request) {
	log.Println("verifying webhook subscription with Strava")
	switch isAppSubscribed() {
	case true:
		log.Printf("yams.energy is already subscribed!\n")
		w.WriteHeader(http.StatusOK)
	case false:
		log.Printf("performing hub challenge with strava...\n")
		var hubMode, hubToken string = r.URL.Query().Get("hub.mode"), r.URL.Query().Get("hub.verify_token")
		if hubMode == "subscribe" && hubToken == os.Getenv("STRAVA_VERIFY_TOKEN") {
			log.Printf("hub challenge passed!\n")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("{ \"hub.challenge\":\"%s\" }", r.URL.Query().Get("hub.challenge"))))
		} else {
			log.Printf("hub challenge failed, verification tokens do not match.\n")
		}
	}
}

func isAppSubscribed() bool {
	params := url.Values{}
	params.Add("client_id", os.Getenv("STRAVA_CLIENT_ID"))
	params.Add("client_secret", os.Getenv("STRAVA_CLIENT_SECRET"))
	resp, err := http.Get("https://www.strava.com/api/v3/push_subscriptions?" + params.Encode())
	if err != nil {
		log.Printf("request failed: %s\n", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	bodyMap := []map[string]interface{}{}
	_ = json.Unmarshal(body, &bodyMap)
	if len(bodyMap) == 0 {
		return false
	} else {
		if bodyMap[0]["id"] != nil {
			return true
		} else {
			return false
		}
	}
}
