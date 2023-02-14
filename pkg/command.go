package pkg

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (handler *MessageHandler) Command(update *tgbotapi.Update) error {
	if !update.Message.IsCommand() { // ignore any non-command Messages
		return nil
	}

	// Create a new MessageConfig. We don't have text yet,
	// so we leave it empty.
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

	switch update.Message.Command() {
	case CMD_OPEN_MENU:
		msg.Text = " 📜 Menu đã được thêm vào"
		if update.Message.Chat.IsPrivate() {
			msg.ReplyMarkup = PrivateKeyboard
		} else {
			msg.ReplyMarkup = LobbyKeyboard
		}
	case CMD_CLOSE_MENU:
		msg.Text = " ❌  Loại bỏ Menu"
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)

		handler.removeMessage(update.Message.Chat.ID, update.Message.MessageID)
	default:
		msg.Text = "Tạm tời em không hiểu. Để em cập nhật thêm sau nhé!"
	}

	handler.sendMessage(msg)

	return nil
}
