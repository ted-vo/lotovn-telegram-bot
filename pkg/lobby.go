package pkg

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aquasecurity/table"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var GameInChatMap = make(map[int64]*Lobby)

type Lobby struct {
	ChatId    int64
	GameId    int
	players   map[int64]*Player
	lifecycle Lifecycle
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
			lifecycle: NewGame(
				time.Second*10,
				TicketConifg{
					MaxNumer:       70,
					MaxRow:         9,
					MaxCol:         7,
					MaxNumberOfRow: 4,
				},
			),
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

	if currentGame.lifecycle.isStarted() {
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
		Ticket:   NewTicket(currentGame.GameId, currentGame.lifecycle.ticketConfig()),
	}
	currentGame.players[registor.ID] = player

	// send ticket for player in Private
	ticketText, _ := Parse("./config/ticket.html",
		struct {
			GameId   int
			TicketId uint32
			Data     string
		}{
			GameId:   currentGame.GameId,
			TicketId: player.Ticket.Id.ID(),
			Data:     "",
		})
	msgPlayer := tgbotapi.NewMessage(
		player.Id,
		ticketText,
	)
	msgPlayer.ParseMode = "HTML"
	msgPlayer.ReplyMarkup = GenerateTicketKeyboard(currentGame.ChatId, currentGame.GameId, player.Ticket.board)
	resMsg := handler.sendMessage(msgPlayer)
	// tracked msg of ticket send to player for clear when game end
	player.Ticket.MessageId = resMsg.MessageID

	handler.updateListPlayerState(currentGame)

	return nil
}

func (handler *MessageHandler) start(update *tgbotapi.Update) error {
	chatId := update.CallbackQuery.Message.Chat.ID
	var currentGame = GameInChatMap[chatId]
	if currentGame == nil {
		return fmt.Errorf("Game không tồn tại. Vui lòng mở báo danh!")
	}

	if currentGame.lifecycle.isStarted() {
		return fmt.Errorf("Game đã bắt đầu rồi mà.")
	}

	msg := tgbotapi.NewMessage(chatId, "Game bắt đầu!")
	msg.ReplyToMessageID = currentGame.GameId
	handler.sendMessage(msg)

	releaseChanel := currentGame.lifecycle.start()
	go func(chatId int64, c chan int, handler *MessageHandler) {
		for {
			res, ok := <-c
			if ok == false {
				break
			}
			currentGame.lifecycle.addResultSeed(res)
			handler.sendMessage(tgbotapi.NewMessage(
				chatId,
				fmt.Sprintf("Số %d", res)))
		}
	}(chatId, releaseChanel, handler)

	handler.updateListPlayerState(currentGame)

	return nil
}

func (handler *MessageHandler) pause(update *tgbotapi.Update) error {
	chatId := update.CallbackQuery.Message.Chat.ID
	var currentGame = GameInChatMap[chatId]
	if currentGame == nil {
		return fmt.Errorf("Game không tồn tại. Vui lòng mở báo danh!")
	}
	if currentGame.lifecycle.isPaused() {
		return fmt.Errorf("Game đã dừng rồi mà.")
	}

	msg := tgbotapi.NewMessage(chatId, "Game tạm dừng!")
	msg.ReplyToMessageID = currentGame.GameId
	handler.sendMessage(msg)

	go currentGame.lifecycle.pause()

	handler.updateListPlayerState(currentGame)

	return nil
}

func (handler *MessageHandler) resume(update *tgbotapi.Update) error {
	chatId := update.CallbackQuery.Message.Chat.ID
	var currentGame = GameInChatMap[chatId]
	if currentGame == nil {
		return fmt.Errorf("Game không tồn tại. Vui lòng mở báo danh!")
	}
	if currentGame.lifecycle.isStarted() {
		return fmt.Errorf("Game đã bắt đầu rồi mà.")
	}

	msg := tgbotapi.NewMessage(chatId, "Game tiếp tục!")
	msg.ReplyToMessageID = currentGame.GameId
	handler.sendMessage(msg)

	currentGame.lifecycle.resume()

	handler.updateListPlayerState(currentGame)

	return nil
}

func (handler *MessageHandler) finish(update *tgbotapi.Update) error {
	chatId := update.CallbackQuery.Message.Chat.ID
	var currentGame = GameInChatMap[chatId]
	if currentGame == nil {
		return fmt.Errorf("Game không tồn tại. Vui lòng mở báo danh!")
	}

	go currentGame.lifecycle.stop()

	handler.updateListPlayerState(currentGame)

	msg := tgbotapi.NewMessage(chatId, "Kết thúc!")
	msg.ReplyToMessageID = currentGame.GameId
	handler.sendMessage(msg)

	// update message ticket for user after game end
	for _, v := range currentGame.players {
		ticketText, _ := Parse(
			"./config/ticket.html",
			struct {
				GameId   int
				TicketId uint32
				Data     string
			}{
				GameId:   v.Ticket.GameId,
				TicketId: v.Ticket.Id.ID(),
				Data:     BeautyTicket(v.Ticket.board),
			})

		editMessage := tgbotapi.NewEditMessageText(
			v.Id,
			v.Ticket.MessageId,
			ticketText,
		)
		editMessage.ParseMode = "HTML"
		handler.editMessage(editMessage)
	}

	return nil
}

func (handler *MessageHandler) wait(update *tgbotapi.Update) error {
	arrData := strings.Split(update.CallbackQuery.Data, ";")
	gameChatId, _ := strconv.ParseInt(arrData[1], 10, 64)

	var currentGame = GameInChatMap[gameChatId]
	if currentGame == nil {
		return fmt.Errorf("Game không tồn tại. Vui lòng mở báo danh!")
	}
	if currentGame.lifecycle.status() == LOBBY {
		return fmt.Errorf("Game chưa bắt đầu. Chờ chút nào!")
	}

	player := currentGame.players[update.CallbackQuery.From.ID]
	player.Wait += 1

	handler.updateListPlayerState(currentGame)

	handler.sendMessage(tgbotapi.NewMessage(
		gameChatId,
		fmt.Sprintf("@%s đợi lần thứ %d", player.Username, player.Wait),
	))

	return nil
}

func (handler *MessageHandler) bingo(update *tgbotapi.Update) error {
	arrData := strings.Split(update.CallbackQuery.Data, ";")
	gameChatId, _ := strconv.ParseInt(arrData[1], 10, 64)

	var currentGame = GameInChatMap[gameChatId]
	if currentGame == nil {
		return fmt.Errorf("Game không tồn tại. Vui lòng mở báo danh!")
	}
	if currentGame.lifecycle.status() == LOBBY {
		return fmt.Errorf("Game chưa bắt đầu. Chờ chút nào!")
	}

	player := currentGame.players[update.CallbackQuery.From.ID]
	text, _ := Parse("./config/bingo.html",
		struct {
			Username string
			TicketId uint32
			GameId   int
			Result   string
			Data     string
		}{
			Username: player.Username,
			TicketId: player.Ticket.Id.ID(),
			GameId:   currentGame.GameId,
			Result:   BeautyResult(currentGame.lifecycle.result()),
			Data:     BeautyTicket(player.Ticket.board),
		})
	msg := tgbotapi.NewMessage(gameChatId, text)
	msg.ParseMode = "HTML"
	msg.ReplyToMessageID = currentGame.GameId
	handler.sendMessage(msg)

	currentGame.lifecycle.pause()

	handler.updateListPlayerState(currentGame)

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

func (handler *MessageHandler) updateListPlayerState(game *Lobby) {
	text, _ := Parse("./config/game.html",
		struct {
			GameId int
			List   string
		}{
			GameId: game.GameId,
			List:   game.renderPlayerList(),
		})

	var inlineKeyboard tgbotapi.InlineKeyboardMarkup
	switch game.lifecycle.status() {
	case STARTED:
		inlineKeyboard = PlayingInnlineKeyboard
	case PAUSED:
		inlineKeyboard = PausedInlineKeyboard
	case LOBBY:
		inlineKeyboard = OpenGameInlineKeyboard
	default:
	}

	var editMsg tgbotapi.EditMessageTextConfig
	if game.lifecycle.status() == STOPPED {
		editMsg = tgbotapi.NewEditMessageText(
			game.ChatId,
			game.GameId,
			text,
		)
	} else {
		editMsg = tgbotapi.NewEditMessageTextAndMarkup(
			game.ChatId,
			game.GameId,
			text,
			inlineKeyboard,
		)
	}
	editMsg.ParseMode = "HTML"

	handler.editMessage(editMsg)
}
