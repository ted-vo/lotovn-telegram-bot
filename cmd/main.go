package main

import (
	"os"

	"github.com/apex/log"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/ted-vo/lotovn-telegram-bot/pkg"
)

func main() {
	log.SetHandler(pkg.NewLogHandler())
	// go startHTTPServer()

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TOKEN"))
	if err != nil {
		log.Error(err.Error())
	}

	bot.Debug = os.Getenv("DEBUG") == "true"

	log.Infof("Authorized on account %s", bot.Self.UserName)

	// Create a new UpdateConfig struct with an offset of 0. Offsets are used
	// to make sure Telegram knows we've handled previous values and we don't
	// need them repeated.
	u := tgbotapi.NewUpdate(0)

	// Tell Telegram we should wait up to 30 seconds on each request for an
	// update. This way we can get information just as quickly as making many
	// frequent requests without having to send nearly as many.
	u.Timeout = 30

	// Start polling Telegram for updates.
	updates := bot.GetUpdatesChan(u)

	// handler := pkg.NewHandler(bot, pkg.GetSheet())
	handler := pkg.NewHandler(bot)

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Infof("[%s] %s", update.Message.From.UserName, update.Message.Text)

			if update.Message.IsCommand() {
				handler.Command(&update)
			} else {
				handler.Keyboard(&update)
			}
		} else if update.CallbackQuery != nil {
			handler.InlineKeyboard(&update)
		}
	}
}
