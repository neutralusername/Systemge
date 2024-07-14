package main

import (
	"Systemge/Error"
	"Systemge/Http"
	"Systemge/Oauth2"
	"Systemge/Utilities"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

func main() {
	randomizer := Utilities.NewRandomizer(Utilities.GetSystemTime())
	oauth2Server, err := (Oauth2.Config{
		Randomizer:              randomizer,
		Oauth2State:             randomizer.GenerateRandomString(16, Utilities.ALPHA_NUMERIC),
		SessionLifetimeMs:       5000,
		Port:                    8081,
		AuthPath:                "/auth",
		AuthCallbackPath:        "/callback",
		SucessCallbackRedirect:  "http://localhost:8080",
		FailureCallbackRedirect: "http://google.com",
		OAuth2Config: &oauth2.Config{
			ClientID:     "1261641608886222908",
			ClientSecret: "xD",
			RedirectURL:  "http://localhost:8081/callback",
			Scopes:       []string{"identify"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://discord.com/api/oauth2/authorize",
				TokenURL: "https://discord.com/api/oauth2/token",
			},
		},
		Logger:       Utilities.NewLogger("test.log", "test.log", "test.log", "test.log"),
		TokenHandler: tokenHandler,
	}).New()
	if err != nil {
		panic(err)
	}
	oauth2Server.Start()

	httpServer := Http.New(8080, map[string]Http.RequestHandler{
		"/": func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
			sessionId := httpRequest.URL.Query().Get("sessionId")
			session := oauth2Server.GetSession(sessionId)
			if session == nil {
				responseWriter.Write([]byte("invalid session"))
				return
			}
			username, ok := session.Get("username")
			if !ok {
				responseWriter.Write([]byte("invalid session"))
				return
			}
			responseWriter.Write([]byte("Hello " + username.(string)))
		},
	})
	Http.Start(httpServer, "", "")

	time.Sleep(1000 * time.Second)
}

func tokenHandler(oauth2Server *Oauth2.Server, token *oauth2.Token) (map[string]interface{}, error) {
	client := oauth2Server.GetOauth2Config().Client(context.Background(), token)
	resp, err := client.Get("https://discord.com/api/users/@me")
	if err != nil {
		return nil, Error.New("failed getting user", err)
	}
	defer resp.Body.Close()

	var discordAuthData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&discordAuthData); err != nil {
		return nil, Error.New("failed decoding user", err)
	}
	return discordAuthData, nil
}
