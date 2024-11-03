package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"html/template"
	"log"
	"net/url"
	"os"
)

func main() {
	renderLandingPage()
	renderAuthFailPage()
	renderAuthPassPage()
	prepareFoodItems()
	log.Println("Finished building!")
}

func renderLandingPage() {
	tmpl, _ := template.ParseFiles("web/templates/index.html")
	html, _ := os.Create("public/index.html")
	_ = tmpl.Execute(html, map[string]string{
		"StravaAuthorizationLink": createStravaLink(),
		"GitSHA":                  os.Getenv("VERCEL_GIT_COMMIT_SHA")[:8],
	})
	_ = html.Close()
}

func createStravaLink() string {
	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("client_id", os.Getenv("STRAVA_CLIENT_ID"))
	params.Add("scope", "read,activity:write,activity:read_all")
	params.Add("approval_prompt", "auto")
	params.Add("redirect_uri", "https://yams.energy/api/strava/authorization")
	return "https://www.strava.com/oauth/authorize?" + params.Encode()
}

func renderAuthPassPage() {
	tmpl, _ := template.ParseFiles("web/templates/authorization-pass.html")
	html, _ := os.Create("public/authorization-pass/index.html")
	_ = tmpl.Execute(html, nil)
	_ = html.Close()
}

func renderAuthFailPage() {
	tmpl, _ := template.ParseFiles("web/templates/authorization-fail.html")
	html, _ := os.Create("public/authorization-fail/index.html")
	_ = tmpl.Execute(html, nil)
	_ = html.Close()
}

func prepareFoodItems() {

	// load binary blob
	gobBytes, err := os.ReadFile("app/fooditems.bin")
	if err != nil {
		log.Println("> error opening food items blob:", err)
	}

	// decrypt
	k, _ := hex.DecodeString(os.Getenv("FOOD_KEY"))
	c, _ := aes.NewCipher(k)
	gcm, _ := cipher.NewGCM(c)
	nonce, ciphertext := gobBytes[:gcm.NonceSize()], gobBytes[gcm.NonceSize():]
	decryptedData, _ := gcm.Open(nil, nonce, ciphertext, nil)

	// save
	err = os.WriteFile("app/fooditems.go", decryptedData, 0600)
	if err != nil {
		panic(err)
	}
}
