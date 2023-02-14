package pkg

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler interface {
	Command(update *tgbotapi.Update) error
	Keyboard(update *tgbotapi.Update) error
	InlineKeyboard(update *tgbotapi.Update) error
}

type MessageHandler struct {
	bot             *tgbotapi.BotAPI
	SpreadsheetClub *SpreadsheetClub
}

//	func NewHandler(bot *tgbotapi.BotAPI, sheetClub *SpreadsheetClub) Handler {
//		return &MessageHandler{
//			bot:             bot,
//			SpreadsheetClub: sheetClub,
//		}
//	}

func NewHandler(bot *tgbotapi.BotAPI) Handler {
	return &MessageHandler{
		bot: bot,
	}
}
