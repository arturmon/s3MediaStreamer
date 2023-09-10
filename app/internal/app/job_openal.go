package app

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"skeleton-golange-application/model"
	"time"

	"github.com/ayush6624/go-chatgpt"
	"github.com/google/uuid"
)

func (e OpenAIJob) Run() {
	e.app.logger.Println("init ChatGPT...")
	radDBdata, err := e.app.storage.Operations.GetAlbumsForLearn()
	if err != nil {
		e.app.logger.Fatal(err)
	}
	e.app.logger.Traceln(radDBdata)

	if e.app.cfg.AppConfig.OpenAI.OpenAiKey == "" {
		e.app.logger.Println("OpenAI key is not provided. Aborting ChatGPT job.")
		return
	}

	var respAlbums []model.Tops
	client, errAi := chatgpt.NewClient(e.app.cfg.AppConfig.OpenAI.OpenAiKey)
	if errAi != nil {
		e.app.logger.Fatal(err)
	}
	ctx := context.Background()

	var titleAndArtist string
	for _, album := range radDBdata {
		titleAndArtist += fmt.Sprintf("Title: %s, Artist: %s,", album.Title, album.Artist)
	}

	res, errSend := client.Send(ctx, &chatgpt.ChatCompletionRequest{
		Model: chatgpt.GPT35Turbo,
		Messages: []chatgpt.ChatMessage{
			{
				Role:    chatgpt.ChatGPTModelRoleAssistant,
				Content: "suggest 10 songs based on this selection, output format: Title: %s, Artist: %s, Description: %s" + titleAndArtist,
			},
		},
	})
	if errSend != nil {
		e.app.logger.Fatal(err)
	}
	a, _ := json.MarshalIndent(res, "", "  ")

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	err = json.Unmarshal(a, &response)
	if err != nil {
		e.app.logger.Errorf("Failed to unmarshal JSON: %v", err)
		return
	}
	re := regexp.MustCompile(`Title: (.*), Artist: (.*), Description: (.*)`)
	for _, choice := range response.Choices {
		matches := re.FindAllStringSubmatch(choice.Message.Content, -1)
		for _, submatches := range matches {
			if len(submatches) < minSubmatchesCount {
				e.app.logger.Errorf("Failed to parse response: %v", err)
				continue
			}
			respAlbum := model.Tops{
				ID:          uuid.New(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				Title:       submatches[1],
				Artist:      submatches[2],
				Description: submatches[3],
				Sender:      "open_ai",
			}
			userUUID, errUUID := uuid.Parse(e.app.cfg.AppConfig.OpenAI.UUIDWriteUser)
			if errUUID != nil {
				e.app.logger.Errorf("Failed to parse UUID: %v", errUUID)
				continue
			}
			respAlbum.CreatorUser = userUUID
			respAlbums = append(respAlbums, respAlbum)
		}
	}
	e.app.logger.Tracef("AI: %+v", respAlbums)
	err = e.app.storage.Operations.CreateTops(respAlbums)
	if err != nil {
		e.app.logger.Fatal(err)
	}

	e.app.logger.Println("complete ChatGPT")
}
