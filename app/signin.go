package app

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleUserId struct {
	Id            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Picture       string `json:"picture"`
}

var googleOauthConfig = oauth2.Config{
	RedirectURL:  "http://localhost:3000/auth/google/callback",
	ClientID:     "818689212305-2mv39kg6l4t1o9hes408bi7uabn51o62.apps.googleusercontent.com",
	ClientSecret: "2uuTl5Y5VTmW1WEPbRsiUSlJ",
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
	Endpoint:     google.Endpoint,
}

func googleLoginHandler(w http.ResponseWriter, r *http.Request) {
	state := generateStateOauthCookie(w)
	url := googleOauthConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func generateStateOauthCookie(w http.ResponseWriter) string {
	expiration := time.Now().Add(1 * 24 * time.Hour)

	b := make([]byte, 16)
	rand.Read(b)

	state := base64.URLEncoding.EncodeToString(b)

	cookie := &http.Cookie{
		Name:    "oauthstate",
		Value:   state,
		Expires: expiration,
	}

	http.SetCookie(w, cookie)
	return state
}

func googleAuthCallback(w http.ResponseWriter, r *http.Request) {
	oauthstate, _ := r.Cookie("oauthstate")
	if r.FormValue("state") != oauthstate.Value {
		errMsg := fmt.Sprintf("invalid google oauth state cookie: %s state: %s\n", oauthstate.Value, r.FormValue("state"))
		log.Printf(errMsg)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	data, err := getGoogleUserInfo(r.FormValue("code"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var userInfo GoogleUserId
	err = json.Unmarshal(data, &userInfo)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session, err := store.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println(userInfo)

	session.Values["id"] = userInfo.Id
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

func getGoogleUserInfo(code string) ([]byte, error) {
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("Failed to Exchange %s\n", err.Error())
	}

	resp, err := http.Get(oauthGoogleUrlAPI + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("Failed to Get userInfo %s\n", err.Error())
	}

	return ioutil.ReadAll(resp.Body)
}
