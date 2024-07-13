package main

import (
	"Systemge/Http"
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
		ClientSecret: "oAihi_Ezi3n-6CpUwE4OSohVxXeQ3vQK",
		RedirectURL:  "http://localhost:8080/callback",
		Scopes:       []string{"identify"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discord.com/api/oauth2/authorize",
			TokenURL: "https://discord.com/api/oauth2/token",
		},
	}
	oauth2State = "randomState" // Generate a random state for each session in production
)

func main() {
	// Print the client ID and secret to verify they are being loaded correctly
	fmt.Printf("Client ID: %s\n", discordOAuth2Config.ClientID)
	fmt.Printf("Client Secret: %s\n", discordOAuth2Config.ClientSecret)

	http.HandleFunc("/", Http.DiscordAuth(discordOAuth2Config, oauth2State))
	http.HandleFunc("/callback", handleCallback)

	log.Println("Server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %v\n", err)
	}
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	log.Println("HandleCallback: Received request")
	state := r.URL.Query().Get("state")
	if state != oauth2State {
		http.Error(w, "invalid oauth state", http.StatusUnauthorized)
		return
	}

	code := r.URL.Query().Get("code")
	token, err := discordOAuth2Config.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("discordOAuth2Config.Exchange() failed with '%s'\n", err)
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	client := discordOAuth2Config.Client(context.Background(), token)
	resp, err := client.Get("https://discord.com/api/users/@me")
	if err != nil {
		log.Printf("client.Get() failed with '%s'\n", err)
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var user map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		log.Printf("json.Decode() failed with '%s'\n", err)
		http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
		return
	}

	// Use user info (e.g., user["id"], user["username"]) to identify the client
	fmt.Printf("User Info: %+v\n", user)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
