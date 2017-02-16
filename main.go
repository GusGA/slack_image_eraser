package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var slackToken string

func init() {
	slackToken = os.Getenv("SLACK_TOKEN")
	if slackToken == "" {
		log.Fatal("You must set a valid environment variable SLACK_TOKEN")
		os.Exit(1)
	}
}

const slackImageURL = "https://slack.com/api/files.list"

type slackObject struct {
	ID           string `json:"id"`
	DisplayAsBot bool   `json:"display_as_bot"`
}

func getImages(page int) (images []*slackObject) {
	const tsTo = 1451520000 // 31-December-2015
	url := fmt.Sprintf("https://slack.com/api/files.list?token=%s&ts_to%d&types=images&page=%d", slackToken, tsTo, page)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)

	var r struct {
		Ok     bool `json:"ok"`
		Paging struct {
			Page  int `json:"page"`
			Pages int `json:"pages"`
			Total int `json:"total"`
		} `json:"paging"`
		Files []*slackObject
	}

	err = json.Unmarshal(data, &r)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(r.Paging.Page)
	fmt.Println(r.Paging.Pages)

	for _, image := range r.Files {
		if !image.DisplayAsBot {
			images = append(images, image)
		}
	}
	if r.Paging.Pages > r.Paging.Page {
		images = append(images, getImages(r.Paging.Page+1)...)
	}

	return images
}

func deleteImages(images []*slackObject) {
	type deleteResponse struct {
		Ok    bool   `json:"ok"`
		Error error  `json:",omitempty"`
		ID    string `json:",omitempty"`
	}

	deleteCh := make(chan *deleteResponse, 20)

	for _, image := range images {
		go func(id string) {
			url := fmt.Sprintf("https://slack.com/api/files.delete?token=%s&file=%s", slackToken, id)
			resp, err := http.Get(url)
			defer resp.Body.Close()
			content := &deleteResponse{Ok: false}
			if err != nil {
				content.Error = err
			} else {
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Fatal(err)
				}
				if err := json.Unmarshal(body, content); err != nil {
					log.Fatal(err)
				}
			}
			content.ID = id
			deleteCh <- content

		}(rem.ID)

		select {
		case content := <-deleteCh:
			if content.Ok {
				fmt.Printf("Image id: %s was successfully deleted\n", content.ID)
			} else {
				fmt.Printf("Image id %s could not be deleted, error %s\n", content.ID, content.Error)
			}

		}
	}
}

func main() {
	images := getImages(1)
	fmt.Println("Deleting %d images"len(images))
	deleteImages(images)
}
