package pkg

import (
	"fmt"
	"strings"

	"github.com/apex/log"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// ------------------------------------------------------------------------
// -------------------------  Private function ----------------------------
// ------------------------------------------------------------------------

func (handler *MessageHandler) getCaller(update *tgbotapi.Update) string {
	caller := fmt.Sprintf("@%s", update.Message.From.UserName)
	if len(caller) == 0 {
		caller = fmt.Sprintf("%s %s", update.Message.From.FirstName, update.Message.From.LastName)
	}

	return caller
}

func (handler *MessageHandler) sendMessage(msg tgbotapi.MessageConfig) *tgbotapi.Message {
	if len(msg.Text) != 0 {
		msg, err := handler.bot.Send(msg)
		if err != nil {
			log.Error(err.Error())
		}
		return &msg
	}

	return nil
}

func (handler *MessageHandler) editMessage(msg tgbotapi.Chattable) (tgbotapi.Message, error) {
	return handler.bot.Send(msg)
}

func (handler *MessageHandler) removeMessage(chatId int64, messageId int) {
	if _, err := handler.bot.Request(tgbotapi.NewDeleteMessage(chatId, messageId)); err != nil {
		log.Errorf("delete message erorr: %s", err.Error())
	}
}

func EscapeSpecialCharacters(original string) string {
	chars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	replacers := make([]string, 0)
	for _, c := range chars {
		replacers = append(replacers, c)
		replacers = append(replacers, fmt.Sprintf("\\%s", c))
	}
	return strings.NewReplacer(replacers...).Replace(original)
}
