package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var config = &oauth2.Config{
	// ClientID:     "604508162253-87ajfsafq6n3b66qnj0mo24dtekicppj.apps.googleusercontent.com",
	// ClientSecret: "",
	Endpoint: google.Endpoint,
	// Scopes:   []string{urlshortener.UrlshortenerScope},
}

func main() {
	credentials, err := ioutil.ReadFile("client_secret_604508162253-87ajfsafq6n3b66qnj0mo24dtekicppj.apps.googleusercontent.com.json")
	if err != nil {
		log.Fatalf("Error reading credentials file: %v", err)
	}

	config, err := google.ConfigFromJSON(
		credentials,
		"https://www.googleapis.com/auth/photoslibrary.readonly",
		"https://www.googleapis.com/auth/devstorage.read_write",
	)
	if err != nil {
		log.Fatalf("Failed to google.ConfigFromJSON: %v\n", err)
	}

	// ctx := context.Background()

	var tok *oauth2.Token

	tokenJsonStr, err := ioutil.ReadFile("token.json")
	if err != nil {
		// TODO
	} else {
		err = json.Unmarshal(tokenJsonStr, &tok)
		if err != nil {
			log.Fatalf("err: %v\n", err)
		}
	}

	if tok == nil {
		url := config.AuthCodeURL("state")
		fmt.Printf("Visit the URL for the auth dialog:\n%v\n", url)

		var code string
		if _, err := fmt.Scan(&code); err != nil {
			log.Fatal(err)
		}

		if err != nil {
			log.Fatalf("Failed to config.Exchange %v\n", err)
		}

		marshal, err := json.Marshal(tok)
		if err != nil {
			log.Fatalf("Failed to json.Marshal: %v\n", err)
		}
		fmt.Println(string(marshal))
	}

	client := config.Client(oauth2.NoContext, tok)

	resp, err := client.Get("https://photoslibrary.googleapis.com/v1/albums")
	if err != nil {
		log.Fatalf("Failed to client.Get: %v\n", err)
	}

	fmt.Println(resp)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to io.ReadAll: %v\n", err)
	}

	fmt.Println(string(body))
	// TODO: Get photos and upload to Google Cloud Storage
}
