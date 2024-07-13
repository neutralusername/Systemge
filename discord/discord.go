package main

import (
	"Systemge/Http"
	"Systemge/Utilities"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/oauth2"
)

var (
	discordOAuth2Config = &oauth2.Config{
		ClientID:     "1261641608886222908",
		ClientSecret: "xD",
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"identify"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discord.com/api/oauth2/authorize",
			TokenURL: "https://discord.com/api/oauth2/token",
		},
	}
	oauth2State = "randomState" // Generate a random state for each session in production
)

var logger = Utilities.NewLogger("test.log", "test.log", "test.log", "test.log")

func main() {
	// Print the client ID and secret to verify they are being loaded correctly
	fmt.Printf("Client ID: %s\n", discordOAuth2Config.ClientID)
	fmt.Printf("Client Secret: %s\n", discordOAuth2Config.ClientSecret)

	tokenChannel := make(chan *oauth2.Token)

	http.HandleFunc("/", Http.DiscordAuth(discordOAuth2Config, oauth2State))
	http.HandleFunc("/callback", Http.DiscordAuthCallback(discordOAuth2Config, oauth2State, logger, tokenChannel))

	go func() {
		token := <-tokenChannel
		println("t")
		handleToken(token)
	}()

	log.Println("Server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %v\n", err)
	}
}

func handleToken(token *oauth2.Token) {
	client := discordOAuth2Config.Client(context.Background(), token)
	resp, err := client.Get("https://discord.com/api/users/@me")
	if err != nil {
		log.Fatalf("Failed getting user: %v\n", err)
	}
	defer resp.Body.Close()

	var user map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		println(err.Error())
		return
	}

	fmt.Printf("User Info: %+v\n", user)
}
