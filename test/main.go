package main

import (
	"Systemge/Error"
	"Systemge/Http"
	"Systemge/Oauth2"
	"Systemge/Utilities"
	"context"
	"encoding/json"
	"time"

	"golang.org/x/oauth2"
)

var logger = Utilities.NewLogger("test.log", "test.log", "test.log", "test.log")

var discordOAuth2Config = &oauth2.Config{
	ClientID:     "1261641608886222908",
	ClientSecret: "xD",
	RedirectURL:  "http://localhost:8081/callback",
	Scopes:       []string{"identify"},
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://discord.com/api/oauth2/authorize",
		TokenURL: "https://discord.com/api/oauth2/token",
	},
}

func main() {
	httpServer := Http.New(8080, map[string]Http.RequestHandler{
		"/": Http.SendHTTPResponseCodeAndBody(200, "Hello World"),
	})
	Http.Start(httpServer, "", "")

	oauth2Server := Oauth2.NewServer(8081, "/auth", "/callback", discordOAuth2Config, logger, tokenHandler)
	oauth2Server.Start()
	time.Sleep(1000 * time.Second)
}

func tokenHandler(oauth2Server *Oauth2.Server, token *oauth2.Token) (map[string]interface{}, error) {
	client := discordOAuth2Config.Client(context.Background(), token)
	resp, err := client.Get("https://discord.com/api/users/@me")
	if err != nil {
		return nil, Error.New("failed getting user", err)
	}
	defer resp.Body.Close()

	var user map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, Error.New("failed decoding user", err)
	}
	return user, nil
}
