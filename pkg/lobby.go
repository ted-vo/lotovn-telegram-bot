package pkg

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var GameInChatMap = make(map[int64]*Lobby)

type Lobby struct {
	ChatId  int64
	GameId  int
	players map[int64]*Player
	IsStart bool
}

func (lobby *Lobby) String() string {
	return fmt.Sprintf("🎯 Chào mừng bà con cô bác đến với Đoàn Lô Tô Ted Vo!!!\n\nGameId: *%d*\nDanh sách người tham gia\n\n", lobby.GameId)
}

type GameControl interface {
	register(update *tgbotapi.Update)
	start(update *tgbotapi.Update)
	finish(update *tgbotapi.Update)
	wait(update *tgbotapi.Update)
	bingo(update *tgbotapi.Update)
	numberCheck(update *tgbotapi.Update)
}

type Player struct {
	Id       int64
	Username string
	Name     string
	Ticket   *Ticket
}

func (handler *MessageHandler) openGame(update *tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "🎯 Chào mừng bà con cô bác đến với Đoàn Lô Tô Ted Vo")
	chatId := update.Message.Chat.ID

	var currentGame = GameInChatMap[chatId]
	if currentGame == nil {
		msg.ReplyMarkup = OpenGameInlineKeyboard
		respMsg := handler.sendMessage(msg)
		currentGame = &Lobby{
			ChatId:  chatId,
			GameId:  respMsg.MessageID,
			players: make(map[int64]*Player),
		}
		GameInChatMap[chatId] = currentGame

		editMessage := tgbotapi.NewEditMessageTextAndMarkup(
			chatId,
			respMsg.MessageID,
			EscapeSpecialCharacters(currentGame.String()),
			OpenGameInlineKeyboard,
		)
		editMessage.ParseMode = "MarkdownV2"

		handler.editMessage(editMessage)
	} else {
		msg.Text = "Ủa alo? Game hiện tại chưa kết thúc mà, nè..."
		msg.ReplyToMessageID = currentGame.GameId
		handler.sendMessage(msg)
	}

	handler.removeMessage(update.Message.Chat.ID, update.Message.MessageID)

	return nil
}

func (handler *MessageHandler) help(update *tgbotapi.Update) error {

	return nil
}

func (handler *MessageHandler) register(update *tgbotapi.Update) error {
	chatId := update.CallbackQuery.Message.Chat.ID
	var currentGame = GameInChatMap[chatId]
	if currentGame == nil {
		return fmt.Errorf("Game không tồn tại. Vui lòng mở báo danh!")
	}

	registor := update.CallbackQuery.From
	if len(registor.UserName) < 5 {
		return fmt.Errorf("Vui lòng cập nhật `username` trước khi báo danh!")
	}

	player := &Player{
		Id:       registor.ID,
		Username: registor.UserName,
		Name:     fmt.Sprintf("%s %s", registor.FirstName, registor.LastName),
		Ticket: NewTicket(
			TicketConifg{
				GameId:         currentGame.GameId,
				MaxRow:         9,
				MaxCol:         9,
				MaxNumberOfRow: 5,
			}),
	}
	currentGame.players[registor.ID] = player

	// send ticket for player in Private
	msgPlayer := tgbotapi.NewMessage(
		player.Id,
		EscapeSpecialCharacters(player.Ticket.String()),
	)
	msgPlayer.ParseMode = "MarkdownV2"
	msgPlayer.ReplyMarkup = GenerateTicketKeyboard(currentGame.ChatId, player.Ticket.GameId, player.Ticket.board)
	handler.sendMessage(msgPlayer)

	// update list player
	text := currentGame.String()
	for _, v := range currentGame.players {
		text += fmt.Sprintf("@%s - TicketId: *%s*\n", v.Username, v.Ticket.Id)
	}
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		chatId,
		update.CallbackQuery.Message.MessageID,
		EscapeSpecialCharacters(text),
		PlayingInnlineKeyboard,
	)
	editMsg.ParseMode = "MarkdownV2"

	handler.editMessage(editMsg)

	return nil
}

func (handler *MessageHandler) start(update *tgbotapi.Update) error {
	// msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

	return nil
}

func (handler *MessageHandler) finish(update *tgbotapi.Update) error {

	return nil
}

func (handler *MessageHandler) wait(update *tgbotapi.Update) error {

	return nil
}

func (handler *MessageHandler) bingo(update *tgbotapi.Update) error {

	return nil
}

func (handler *MessageHandler) queryNumerCheck(update *tgbotapi.Update) error {
	chatId := update.CallbackQuery.Message.Chat.ID

	arrData := strings.Split(update.CallbackQuery.Data, ";")
	gameChatId, _ := strconv.ParseInt(arrData[1], 10, 64)
	gameId, _ := strconv.Atoi(arrData[2])

	coordinate := strings.Split(arrData[3], "-")
	x, _ := strconv.Atoi(coordinate[0])
	y, _ := strconv.Atoi(coordinate[1])

	var currentGame = GameInChatMap[gameChatId]
	if currentGame == nil {
		return fmt.Errorf("Game không tồn tại. Vui lòng mở báo danh!")
	}

	player := currentGame.players[update.CallbackQuery.From.ID]
	if player == nil {
		return fmt.Errorf("Game không tồn tại. Vui lòng mở báo danh!")
	}

	currentValue := player.Ticket.board[x][y]
	if currentValue == 0 {
		player.Ticket.board[x][y] = -1
	} else if currentValue == -1 {
		player.Ticket.board[x][y] = 0
	} else {
		return nil
	}

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		chatId,
		update.CallbackQuery.Message.MessageID,
		player.Ticket.String(),
		GenerateTicketKeyboard(gameChatId, gameId, player.Ticket.board),
	)
	handler.editMessage(editMsg)

	return nil
}
