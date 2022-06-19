package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/c4pt0r/log"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	req "github.com/imroc/req/v3"
)

var (
	memeAPI = "https://meme-api.herokuapp.com/gimme/ProgrammerHumor"
)

func popMeme() ([]byte, error) {
	client := req.C(). // Use C() to create a client and set with chainable client settings.
				SetTimeout(5 * time.Second).
				DevMode()
	resp, err := client.R(). // Use R() to create a request and set with chainable request settings.
					SetHeader("Accept", "application/vnd.github.v3+json").
					EnableDump(). // Enable dump at request level to help troubleshoot, log content only when an unexpected exception occurs.
					Get(memeAPI)
	if err != nil {
		return nil, err
	}
	if resp.IsSuccess() {
		return resp.Bytes(), nil
	}
	return nil, nil
}

func isMention(update *tgbotapi.Update, who string) bool {
	for _, e := range update.Message.Entities {
		if e.Type == "mention" {
			name := update.Message.Text[e.Offset+1 : e.Offset+e.Length]
			if name == who {
				return true
			}
		}
	}
	return false
}

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		panic(err)
	}

	bot.Debug = true
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	me, err := bot.GetMe()
	if err != nil {
		panic(err)
	}
	log.Infof("I am %s", me.String())

	updates := bot.GetUpdatesChan(updateConfig)
	for update := range updates {
		if update.Message == nil {
			continue
		}

		if !isMention(&update, me.String()) {
			continue
		}

		memeJson, err := popMeme()
		if err != nil {
			log.Errorf("Error: %s", err)
			continue
		}

		var meme map[string]interface{}
		err = json.Unmarshal(memeJson, &meme)
		if err != nil {
			log.Errorf("Error: %s", err)
			continue
		}
		log.D(string(memeJson))
		picURL := meme["url"].(string)
		title := meme["title"].(string)
		link := meme["postLink"].(string)

		file := tgbotapi.FileURL(picURL)
		msg := tgbotapi.NewPhoto(update.Message.Chat.ID, file)
		msg.ReplyToMessageID = update.Message.MessageID
		msg.Caption = fmt.Sprintf("[%s](%s)", title, link)
		msg.ParseMode = "Markdown"

		if _, err := bot.Send(msg); err != nil {
			log.E(err)
		}
	}

}
