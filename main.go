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
	Albums        []Album `json:"albums"`
	NextPageToken string  `json:"nextPageToken"`
}

type SharedAlbums struct {
	Albums        []Album `json:"sharedAlbums"`
	NextPageToken string  `json:"nextPageToken"`
}

type Album struct {
	Id                    string `json:"id"`
	Title                 string `json:"title"`
	ProductUrl            string `json:"productUrl"`
	MediaItemsCount       string `json:"mediaItemsCount"`
	CoverPhotoBaseUrl     string `json:"coverPhotoBaseUrl"`
	CoverPhotoMediaItemId string `json:"coverPhotoMediaItemId"`
}

type MediaItems struct {
	MediaItems []MediaItem `json:"mediaItems"`
}

type MediaItem struct {
	Id              string          `json:"id"`
	Description     string          `json:"description"`
	ProductUrl      string          `json:"productUrl"`
	BaseUrl         string          `json:"baseUrl"`
	MimeType        string          `json:"mimeType"`
	MediaMetadata   MediaMetadata   `json:"mediaMetadata"`
	ContributorInfo ContributorInfo `json:"contributorInfo"`
	Filename        string          `json:"filename"`
}

type MediaMetadata struct{}   // TODO
type ContributorInfo struct{} // TODO

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Missing album title\n")
	}

	argAlbumTitle := os.Args[1]

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

	// Private albums
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

	// fmt.Println(string(body))

	var albums Albums
	err = json.Unmarshal(body, &albums)

	if err != nil {
		log.Fatalf("Failed to json.Unmarshal: %v\n", err)
	}

	// fmt.Printf("%+v\n", albums)
	pretty, _ := json.MarshalIndent(albums, "", "  ")
	fmt.Println(string(pretty))

	// Shared albums
	resp, err = client.Get("https://photoslibrary.googleapis.com/v1/sharedAlbums")
	if err != nil {
		if strings.Contains(err.Error(), "invalid_grant") {
			log.Fatalf("Invalid grant. Please remove your token.json and try again.")
		}
		log.Fatalf("Failed to client.Get: %v\n", err)
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to io.ReadAll: %v\n", err)
	}

	// fmt.Println(string(body))

	var sharedAlbums SharedAlbums
	err = json.Unmarshal(body, &sharedAlbums)

	if err != nil {
		log.Fatalf("Failed to json.Unmarshal: %v\n", err)
	}

	// fmt.Printf("%+v\n", albums)
	pretty, _ = json.MarshalIndent(sharedAlbums, "", "  ")
	fmt.Println(string(pretty))

	// Get photos
	var album *Album

	for _, a := range albums.Albums {
		if a.Title == argAlbumTitle {
			album = &a
		}
	}

	// TODO: Search also from sharedAlbums

	if album == nil {
		log.Fatalf("Failed to find album %+v\n", album)
	}

	// Get photos of the album
	resp, err = client.Post(
		"https://photoslibrary.googleapis.com/v1/mediaItems:search",
		"application/json",
		strings.NewReader(fmt.Sprintf(`{"albumId":"%v"}`, album.Id)))
	if err != nil {
		if strings.Contains(err.Error(), "invalid_grant") {
			log.Fatalf("Invalid grant. Please remove your token.json and try again.")
		}
		log.Fatalf("Failed to client.Get: %v\n", err)
	}

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to io.ReadAll: %v\n", err)
	}

	fmt.Println(string(body))

	var mediaItems MediaItems
	err = json.Unmarshal(body, &mediaItems)

	if err != nil {
		log.Fatalf("Failed to json.Unmarshal: %v\n", err)
	}

	// fmt.Printf("%+v\n", albums)
	pretty, _ = json.MarshalIndent(mediaItems, "", "  ")
	fmt.Println(string(pretty))

	// TODO: upload to Google Cloud Storage
}

// let b:quickrun_config = {'args': 'TMP'}
