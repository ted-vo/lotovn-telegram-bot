package pkg

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/aquasecurity/table"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var GameInChatMap = make(map[int64]*Lobby)

type Lobby struct {
	ChatId  int64
	GameId  int
	players map[int64]*Player
	IsStart bool
}

func (lobby *Lobby) renderPlayerList() string {
	buf := new(bytes.Buffer)
	tb := table.New(buf)
	tb.SetHeaders("STT", "Username", "Mã vé", "Hò")

	i := 1
	for _, player := range lobby.players {
		tb.AddRow(
			fmt.Sprint(i),
			fmt.Sprintf("%s", player.Username),
			fmt.Sprint(player.Ticket.Id.ID()),
			fmt.Sprint(player.Wait),
		)
		i++
	}
	tb.Render()

	return buf.String()
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
	Wait     int
	Ticket   *Ticket
}

func (handler *MessageHandler) openGame(update *tgbotapi.Update) error {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	chatId := update.Message.Chat.ID

	var currentGame = GameInChatMap[chatId]
	if currentGame == nil {
		msg.ReplyMarkup = OpenGameInlineKeyboard
		msg.Text = "🎯 Chào mừng bà con cô bác đến với Đoàn Lô Tô Ted Vo!"
		msg.ParseMode = "HTML"
		respMsg := handler.sendMessage(msg)
		currentGame = &Lobby{
			ChatId:  chatId,
			GameId:  respMsg.MessageID,
			players: make(map[int64]*Player),
		}
		GameInChatMap[chatId] = currentGame

		text, _ := Parse("./config/game.html",
			struct {
				GameId int
				List   string
			}{
				GameId: currentGame.GameId,
				List:   currentGame.renderPlayerList(),
			})
		editMessage := tgbotapi.NewEditMessageTextAndMarkup(
			chatId,
			respMsg.MessageID,
			text,
			OpenGameInlineKeyboard,
		)
		editMessage.ParseMode = "HTML"

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

	if currentGame.IsStart {
		return fmt.Errorf("Game đã bắt đầu. Hãy đợi lượt kế tiếp!")
	}

	registor := update.CallbackQuery.From
	if len(registor.UserName) < 5 {
		return fmt.Errorf("Vui lòng cập nhật `username` trước khi báo danh!")
	}

	if existed := currentGame.players[registor.ID]; existed != nil {
		return fmt.Errorf("@%s > Báo danh rồi thì ngồi im đi nào!", existed.Username)
	}

	player := &Player{
		Id:       registor.ID,
		Username: registor.UserName,
		Name:     fmt.Sprintf("%s %s", registor.FirstName, registor.LastName),
		Ticket: NewTicket(
			TicketConifg{
				GameId:         currentGame.GameId,
				MaxRow:         9,
				MaxCol:         7,
				MaxNumberOfRow: 4,
			}),
	}
	currentGame.players[registor.ID] = player

	// send ticket for player in Private
	ticketText, _ := Parse("./config/ticket.html",
		struct {
			GameId   int
			TicketId uint32
		}{
			GameId:   currentGame.GameId,
			TicketId: player.Ticket.Id.ID(),
		})
	msgPlayer := tgbotapi.NewMessage(
		player.Id,
		ticketText,
	)
	msgPlayer.ParseMode = "HTML"
	msgPlayer.ReplyMarkup = GenerateTicketKeyboard(currentGame.ChatId, player.Ticket.Config.GameId, player.Ticket.board)
	handler.sendMessage(msgPlayer)

	// update list player
	text, _ := Parse("./config/game.html",
		struct {
			GameId int
			List   string
		}{
			GameId: currentGame.GameId,
			List:   currentGame.renderPlayerList(),
		})

	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		chatId,
		update.CallbackQuery.Message.MessageID,
		text,
		OpenGameInlineKeyboard,
	)
	editMsg.ParseMode = "HTML"

	handler.editMessage(editMsg)

	return nil
}

func (handler *MessageHandler) start(update *tgbotapi.Update) error {
	// msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	chatId := update.CallbackQuery.Message.Chat.ID
	var currentGame = GameInChatMap[chatId]
	if currentGame == nil {
		return fmt.Errorf("Game không tồn tại. Vui lòng mở báo danh!")
	}

	if currentGame.IsStart {
		return fmt.Errorf("Game đã bắt đầu rồi mà.")
	}

	// currentGame.Start()

	return nil
}

func (handler *MessageHandler) Pause(update *tgbotapi.Update) error {
	chatId := update.CallbackQuery.Message.Chat.ID
	var currentGame = GameInChatMap[chatId]
	if currentGame == nil {
		return fmt.Errorf("Game không tồn tại. Vui lòng mở báo danh!")
	}

	// currentGame.Pause()
	return nil
}

func (handler *MessageHandler) Resume(update *tgbotapi.Update) error {
	chatId := update.CallbackQuery.Message.Chat.ID
	var currentGame = GameInChatMap[chatId]
	if currentGame == nil {
		return fmt.Errorf("Game không tồn tại. Vui lòng mở báo danh!")
	}

	// currentGame.Resume()
	return nil
}

func (handler *MessageHandler) finish(update *tgbotapi.Update) error {
	chatId := update.CallbackQuery.Message.Chat.ID
	var currentGame = GameInChatMap[chatId]
	if currentGame == nil {
		return fmt.Errorf("Game không tồn tại. Vui lòng mở báo danh!")
	}

	// currentGame.finish()
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

	ticketText, _ := Parse("./config/ticket.html",
		struct {
			GameId   int
			TicketId uint32
		}{
			GameId:   currentGame.GameId,
			TicketId: player.Ticket.Id.ID(),
		})
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(
		chatId,
		update.CallbackQuery.Message.MessageID,
		ticketText,
		GenerateTicketKeyboard(gameChatId, gameId, player.Ticket.board),
	)
	editMsg.ParseMode = "HTML"

	handler.editMessage(editMsg)

	return nil
}
