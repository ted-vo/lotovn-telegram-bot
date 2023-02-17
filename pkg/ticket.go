package pkg

import (
	"bytes"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/aquasecurity/table"
	"github.com/google/uuid"
)

const (
	ROW_SIZE       = 9
	COLUMN_SIZE    = 9
	NUMBER_PER_ROW = 5
)

type Ticket struct {
	Id     uuid.UUID
	GameId int
	Config TicketConifg
	board  [][]int

	lock sync.RWMutex
}

type TicketConifg struct {
	MaxNumer       int
	MaxRow         int
	MaxCol         int
	MaxNumberOfRow int
}

type None struct{}

func NewTicket(gameId int, config TicketConifg) *Ticket {
	ticket := &Ticket{
		Id:     uuid.New(),
		GameId: gameId,
		Config: config,
	}

	for i := 0; i < ticket.Config.MaxRow; i++ {
		ticket.board = append(ticket.board, make([]int, 0))
		for j := 0; j < ticket.Config.MaxCol; j++ {
			ticket.board[i] = append(ticket.board[i], 0)
		}
	}

	ticket.generateNumbers()

	return ticket
}

func randomNumbersInRange(n int, min, max int) []int {
	var indexs []int
	rand.Seed(time.Now().UnixNano())

	historyMap := map[int]*None{}
	for i := 0; i < n; i++ {
		randomNumber := rand.Intn(max-min) + min
		if isExisted := historyMap[randomNumber] != nil; isExisted {
			// if existed num so rollback and re-random
			i--
			continue
		}
		historyMap[randomNumber] = &None{}
		indexs = append(indexs, randomNumber)
	}

	return indexs
}

func shuffleAndPop(values []int) (int, []int) {
	t := time.Now()
	rand.Seed(int64(t.Nanosecond()))

	for i := range values {
		j := rand.Intn(i + 1)
		if i != j {
			values[i], values[j] = values[j], values[i]
		}
	}

	top := len(values) - 1
	value := values[top]
	shuffleValues := values[:top]

	return value, shuffleValues
}

func (ticket *Ticket) generateNumbers() {
	ticket.lock.Lock()
	defer ticket.lock.Unlock()

	for i := 0; i < ticket.Config.MaxRow; i++ {
		min := ticket.Config.MaxNumberOfRow
		max := ticket.Config.MaxCol

		rowsIndex := randomNumbersInRange(
			min,
			0,
			max,
		)

		for _, j := range rowsIndex {
			ticket.board[i][j] = -1
		}
	}

	for i := 0; i < ticket.Config.MaxCol; i++ {
		randomValues := getSeedByIndex(i)

		for j := 0; j < ticket.Config.MaxRow; j++ {
			baseValue := ticket.board[j][i]
			// skip zero value
			if baseValue == 0 {
				continue
			}

			value, shuffleValues := shuffleAndPop(randomValues)
			ticket.board[j][i] = value
			randomValues = shuffleValues
		}
	}
}

func getSeedByIndex(index int) []int {
	var values []int
	if index == 0 {
		for i := 1; i <= 9; i++ {
			values = append(values, i)
		}
	} else if index > 0 && index < 8 {
		for i := index * 10; i <= index*10+9; i++ {
			values = append(values, i)
		}
	} else {
		for i := 80; i <= 90; i++ {
			values = append(values, i)
		}
	}

	return values
}

func BeautyTicket(ticket [][]int) string {
	buf := new(bytes.Buffer)
	tb := table.New(buf)
	tb.SetPadding(1)
	for _, colValues := range ticket {
		var converts []string
		for _, v := range colValues {
			var value string
			switch v {
			case -1:
				value = "*"
			case 0:
				value = ""
			default:
				value = fmt.Sprintf("%d", v)
			}
			converts = append(converts, value)
		}
		tb.AddRows(converts)
	}
	tb.Render()

	return buf.String()
}
