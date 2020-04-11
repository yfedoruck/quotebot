package web

import (
	"encoding/json"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/yfedoruck/quotebot/pkg/env"
	"github.com/yfedoruck/quotebot/pkg/fail"
	"os"
	"path/filepath"
)

type config struct {
	TelegramBotToken string
}

func Token() string {
	file, err := os.Open(env.BasePath() + filepath.FromSlash("/config.json"))
	fail.Check(err)

	decoder := json.NewDecoder(file)
	configuration := config{}

	err = decoder.Decode(&configuration)
	fail.Check(err)

	return configuration.TelegramBotToken
}

func Updates(bot *tgbotapi.BotAPI) tgbotapi.UpdatesChannel {
	if os.Getenv("USERDOMAIN") == "localhost" {
		return longPooling(bot)
	} else {
		return webHooks(bot)
	}
}

// long pooling for localhost
func longPooling(bot *tgbotapi.BotAPI) tgbotapi.UpdatesChannel {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	_, err := bot.RemoveWebhook()
	fail.Check(err)

	updates, err := bot.GetUpdatesChan(u)
	fail.Check(err)

	return updates
}

// web hooks for awake heroku from idling
func webHooks(bot *tgbotapi.BotAPI) tgbotapi.UpdatesChannel {

	conf := tgbotapi.NewWebhook("https://api.telegram.org/bot" + bot.Token + "/setWebhook?url=https://antic-quotes-bot.herokuapp.com/bot" + bot.Token)
	_, err := bot.SetWebhook(conf)
	fail.Check(err)

	return bot.ListenForWebhook("/" + bot.Token)
}
