package pkg

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	TITLE     = "Báo danh mua vé Lô Tô Cô/Chú/Bác/Anh/Chị/Em ơi"
	OPEN_GAME = "📝 Mở báo danh"
	HELP      = "❓ Help"
	FEEDBACK  = "💡 Feedback"

	CMD_OPEN_MENU  = "open"
	CMD_CLOSE_MENU = "close"

	ILB_REGISTER = "🎮 Báo danh"
	ILB_START    = "🎬 Bắt đầu"
	ILB_PAUSE    = "🎬 Tạm dừng"
	ILB_RESUME   = "🎬 Tiếp tục"
	ILB_STOP     = "🎬 Kết thúc"
	ILB_WAIT     = "💣 Hò"
	ILB_BINGO    = "🎊 Kinh"

	QUERY_DATA_REGISTER = "query_register"
	QUERY_DATA_START    = "query_start"
	QUERY_DATA_PAUSE    = "query_pause"
	QUERY_DATA_RESUME   = "query_resume"
	QUERY_DATA_STOP     = "query_stop"
	QUERY_DATA_WAIT     = "query_wait"
	QUERY_DATA_BINGO    = "query_bingo"
	QUERY_DATA_CHECKED  = "query_checked"
)

var LobbyKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(OPEN_GAME),
	),
)

var PrivateKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Hello"),
	),
)

var OpenGameInlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(ILB_REGISTER, QUERY_DATA_REGISTER),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(ILB_START, QUERY_DATA_START),
	),
)

var PlayingInnlineKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(ILB_PAUSE, QUERY_DATA_PAUSE),
		tgbotapi.NewInlineKeyboardButtonData(ILB_STOP, QUERY_DATA_STOP),
	),
)

type Command interface {
	openKeyboard(update *tgbotapi.Update)
	closeKeyboard(update *tgbotapi.Update)
	help(update *tgbotapi.Update)
}

func (handler *MessageHandler) Keyboard(update *tgbotapi.Update) error {
	switch update.Message.Text {
	case OPEN_GAME:
		handler.openGame(update)
	case HELP:
		handler.help(update)
	}

	return nil
}

func (handler *MessageHandler) InlineKeyboard(update *tgbotapi.Update) error {
	// Respond to the callback query, telling Telegram to show the user
	// a message with the data received.
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
	if _, err := handler.bot.Request(callback); err != nil {
		panic(err)
	}

	switch update.CallbackQuery.Data {
	case QUERY_DATA_REGISTER:
		if err := handler.register(update); err != nil {
			text := fmt.Sprintf("Hey %s => %s", getQuerier(update.CallbackQuery.From), err.Error())
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)
			handler.sendMessage(msg)
		}
	case QUERY_DATA_START:
	case QUERY_DATA_WAIT:
	case QUERY_DATA_BINGO:
	default:
		if strings.HasPrefix(update.CallbackQuery.Data, QUERY_DATA_CHECKED) {
			if err := handler.queryNumerCheck(update); err != nil {
				text := fmt.Sprintf("Hey %s => %s", getQuerier(update.CallbackQuery.From), err.Error())
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)
				handler.sendMessage(msg)
			}
		}
	}

	return nil
}

func getQuerier(from *tgbotapi.User) string {
	var name string
	if len(from.UserName) > 5 {
		name = fmt.Sprintf("@%s", from.UserName)
	} else {
		name = fmt.Sprintf("%s %s", from.FirstName, from.LastName)
	}

	return name
}

func GenerateTicketKeyboard(chatId int64, gameId int, board [][]int) tgbotapi.InlineKeyboardMarkup {
	var keyboard [][]tgbotapi.InlineKeyboardButton
	for i, r := range board {
		var row []tgbotapi.InlineKeyboardButton
		for j, number := range r {
			var label string
			var data string

			if number > 0 {
				label = fmt.Sprintf("%d", number)
				data = " "
			} else {
				data = fmt.Sprintf("%s;%d;%d;%d-%d", QUERY_DATA_CHECKED, chatId, gameId, i, j)
				if number < 0 {
					label = "✅"
				} else {
					label = " "
				}
			}

			row = append(row, tgbotapi.NewInlineKeyboardButtonData(label, data))
		}
		keyboard = append(keyboard, row)
	}
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(ILB_WAIT, QUERY_DATA_WAIT),
	))
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(ILB_BINGO, QUERY_DATA_BINGO),
	))

	return tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: keyboard,
	}
}
