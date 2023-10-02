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
	e.app.Logger.Println("init ChatGPT...")
	radDBdata, err := e.app.Storage.Operations.GetAlbumsForLearn()
	if err != nil {
		e.app.Logger.Fatal(err)
	}
	e.app.Logger.Traceln(radDBdata)

	if e.app.Cfg.AppConfig.Jobs.OpenAiKey == "" {
		e.app.Logger.Println("OpenAI key is not provided. Aborting ChatGPT job.")
		return
	}

	var respAlbums []model.Tops
	client, errAi := chatgpt.NewClient(e.app.Cfg.AppConfig.Jobs.OpenAiKey)
	if errAi != nil {
		e.app.Logger.Fatal(err)
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
		e.app.Logger.Fatal(err)
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
		e.app.Logger.Errorf("Failed to unmarshal JSON: %v", err)
		return
	}
	re := regexp.MustCompile(`Title: (.*), Artist: (.*), Description: (.*)`)

	systemUser, errSystemUser := e.app.Storage.Operations.FindUser(e.app.Cfg.AppConfig.Jobs.SystemWriteUser, "email")
	if errSystemUser != nil {
		e.app.Logger.Println("Error find system user")
		return
	}

	parsedUUID, errUUID := uuid.Parse(systemUser.ID.String())
	if errUUID != nil {
		e.app.Logger.Println("Error parsing system user uuid")
		return
	}

	for _, choice := range response.Choices {
		matches := re.FindAllStringSubmatch(choice.Message.Content, -1)
		for _, submatches := range matches {
			if len(submatches) < minSubmatchesCount {
				e.app.Logger.Errorf("Failed to parse response: %v", err)
				continue
			}
			respAlbum := model.Tops{
				ID:          uuid.New(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				Title:       submatches[1],
				Artist:      submatches[2],
				Description: submatches[3],
				Sender:      systemUser.Name,
			}
			respAlbum.CreatorUser = parsedUUID
			respAlbums = append(respAlbums, respAlbum)
		}
	}
	e.app.Logger.Tracef("AI: %+v", respAlbums)
	err = e.app.Storage.Operations.CreateTops(respAlbums)
	if err != nil {
		e.app.Logger.Fatal(err)
	}

	e.app.Logger.Println("complete ChatGPT")
}
