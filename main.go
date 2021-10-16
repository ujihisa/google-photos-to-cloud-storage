package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Albums struct {
	Albums        []Album `albums`
	NextPageToken string  `nextPageToken`
}

type Album struct {
	Id                    string `id`
	Title                 string `title`
	ProductUrl            string `productUrl`
	MediaItemsCount       string `mediaItemsCount`
	CoverPhotoBaseUrl     string `coverPhotoBaseUrl`
	CoverPhotoMediaItemId string `coverPhotoMediaItemId`
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
		// That's fine, it just misses &tok
	} else {
		err = json.Unmarshal(tokenJsonStr, &tok)
		if err != nil {
			log.Fatalf("err: %v\n", err)
		}
	}

	if tok == nil {
		url := config.AuthCodeURL("state")
		fmt.Printf("Visit the URL for the auth dialog:\n%v\n", url)

		cmd := exec.Command("xdg-open", url)
		if err := cmd.Run(); err != nil {
			// Ignore xdg-open failures
		}

		var code string
		if _, err := fmt.Scan(&code); err != nil {
			log.Fatal(err)
		}

		tok, err = config.Exchange(oauth2.NoContext, code)

		if err != nil {
			log.Fatalf("Failed to config.Exchange %v\n", err)
		}

		marshal, err := json.Marshal(tok)
		if err != nil {
			log.Fatalf("Failed to json.Marshal: %v\n", err)
		}

		fmt.Println(string(marshal))

		f, err := os.Create("./token.json")

		if err != nil {
			log.Fatalf("Failed to os.Create: %v\n", err)
		}

		defer f.Close()

		_, err = f.Write(marshal)
		if err != nil {
			log.Fatalf("Failed to f.Write: %v\n", err)
		}
	}

	client := config.Client(oauth2.NoContext, tok)

	resp, err := client.Get("https://photoslibrary.googleapis.com/v1/albums")
	if err != nil {
		if strings.Contains(err.Error(), "invalid_grant") {
			log.Fatalf("Invalid grant. Please remove your token.json and try again.")
		}
		log.Fatalf("Failed to client.Get: %v\n", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to io.ReadAll: %v\n", err)
	}

	fmt.Println(string(body))

	var albums Albums
	err = json.Unmarshal(body, &albums)

	if err != nil {
		log.Fatalf("Failed to json.Unmarshal: %v\n", err)
	}

	fmt.Printf("%#v\n", albums)
	fmt.Println(albums)
	// TODO: It's still somewhat broken...

	// TODO: Get photos and upload to Google Cloud Storage
}
