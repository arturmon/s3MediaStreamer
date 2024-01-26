package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"skeleton-golange-application/app/model"
	"time"

	"github.com/ayush6624/go-chatgpt"
	"github.com/google/uuid"
)

func (e OpenAIJob) Run() {
	e.app.Logger.Println("init ChatGPT...")
	radDBdata, err := e.app.Storage.Operations.GetTracksForLearn()
	if err != nil {
		e.app.Logger.Fatal(err)
	}
	e.app.Logger.Traceln(radDBdata)

	if e.app.Cfg.AppConfig.Jobs.OpenAiKey == "" {
		e.app.Logger.Println("OpenAI key is not provided. Aborting ChatGPT job.")
		return
	}

	var respTracks []model.Tops
	client, errAi := chatgpt.NewClient(e.app.Cfg.AppConfig.Jobs.OpenAiKey)
	if errAi != nil {
		e.app.Logger.Fatal(err)
	}
	ctx := context.Background()

	var titleAndArtist string
	for _, track := range radDBdata {
		titleAndArtist += fmt.Sprintf("Title: %s, Artist: %s,", track.Title, track.Artist)
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
			respTrack := model.Tops{
				ID:          uuid.New(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				Title:       submatches[1],
				Artist:      submatches[2],
				Description: submatches[3],
				Sender:      systemUser.Name,
			}
			respTrack.CreatorUser = parsedUUID
			respTracks = append(respTracks, respTrack)
		}
	}
	e.app.Logger.Tracef("AI: %+v", respTracks)
	err = e.app.Storage.Operations.CreateTops(respTracks)
	if err != nil {
		e.app.Logger.Fatal(err)
	}

	e.app.Logger.Println("complete ChatGPT")
}
