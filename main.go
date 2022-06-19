// mememe
// Copyright (C) mememe author 2022
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"time"

	"github.com/c4pt0r/log"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	req "github.com/imroc/req/v3"
)

var (
	token   = flag.String("tgbot-token", "", "Telegram bot token")
	debug   = flag.Bool("debug", false, "Enable debug mode")
	memeAPI = flag.String("meme-api", "https://meme-api.herokuapp.com/gimme/ProgrammerHumor", "Meme API")
)

func popMeme() ([]byte, error) {
	client := req.C().
		SetTimeout(5 * time.Second)
	resp, err := client.R().
		SetHeader("Accept", "application/vnd.github.v3+json").
		EnableDump().
		Get(*memeAPI)
	if err != nil {
		return nil, err
	}
	if resp.IsSuccess() {
		return resp.Bytes(), nil
	}
	return nil, nil
}

func isGifURL(url string) bool {
	return url[len(url)-4:] == ".gif"
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

func isPrivateMessage(update *tgbotapi.Update) bool {
	return update.Message.Chat.Type == "private"
}

func main() {
	flag.Parse()

	if *token == "" {
		log.Fatal("Telegram bot token is required")
	}

	bot, err := tgbotapi.NewBotAPI(*token)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = *debug
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	me, err := bot.GetMe()
	if err != nil {
		log.Fatal(err)
	}

	updates := bot.GetUpdatesChan(updateConfig)
	for update := range updates {
		if (update.Message == nil) ||
			(!isPrivateMessage(&update) && !isMention(&update, me.String())) {
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
		picURL := meme["url"].(string)
		title := meme["title"].(string)
		link := meme["postLink"].(string)

		file := tgbotapi.FileURL(picURL)
		if isGifURL(picURL) {
			msg := tgbotapi.NewAnimation(update.Message.Chat.ID, file)
			msg.ReplyToMessageID = update.Message.MessageID
			msg.Caption = fmt.Sprintf("[%s](%s)", title, link)
			msg.ParseMode = "Markdown"
			bot.Send(msg)

		} else {
			msg := tgbotapi.NewPhoto(update.Message.Chat.ID, file)
			msg.ReplyToMessageID = update.Message.MessageID
			msg.Caption = fmt.Sprintf("[%s](%s)", title, link)
			msg.ParseMode = "Markdown"
			bot.Send(msg)
		}
	}
}
