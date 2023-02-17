package pkg

import (
	"math/rand"
	"sync"
	"time"
)

type GameStatus int

const (
	LOBBY   GameStatus = 0
	STARTED            = 1
	PAUSED             = 2
	STOPPED            = 3
)

type Lifecycle interface {
	start() chan int
	stop()
	pause()
	resume()
	status() GameStatus
	isStarted() bool
	isPaused() bool
	ticketConfig() TicketConifg
	result() []int
	addResultSeed(number int)
}

type Game struct {
	Status        GameStatus
	Interval      time.Duration
	TicketConifg  TicketConifg
	seed          Seed
	resultSeed    Seed
	ReleaseChanel chan int
	QuitChanel    chan bool
}

type Seed struct {
	numbers []int

	lock sync.RWMutex
}

func NewGame(interval time.Duration, ticketConfig TicketConifg) Lifecycle {
	return &Game{
		Interval:      interval,
		TicketConifg:  ticketConfig,
		seed:          Seed{},
		ReleaseChanel: make(chan int),
		QuitChanel:    make(chan bool),
	}
}

func (seed *Seed) shuffle() {
	t := time.Now()
	rand.Seed(int64(t.Nanosecond()))

	for i := range seed.numbers {
		j := rand.Intn(i + 1)
		if i != j {
			seed.numbers[i], seed.numbers[j] = seed.numbers[j], seed.numbers[i]
		}
	}
}

func (seed *Seed) init(maxNumber int) {
	seed.lock.Lock()
	defer seed.lock.Unlock()

	for i := 1; i <= maxNumber; i++ {
		seed.numbers = append(seed.numbers, i)
	}
}

func (seed *Seed) pop() int {
	seed.lock.Lock()
	defer seed.lock.Unlock()

	seed.shuffle()

	top := len(seed.numbers) - 1
	value := seed.numbers[top]
	seed.numbers = seed.numbers[:top]
	return value
}

func (seed *Seed) push(number int) {
	seed.lock.Lock()
	defer seed.lock.Unlock()

	seed.numbers = append(seed.numbers, number)
}

func (seed *Seed) autoRelease(duration time.Duration, releaseChanel chan int, quit chan bool) {
	for {
		select {
		case <-quit:
			return
		default:
			value := seed.pop()
			releaseChanel <- value

			if len(seed.numbers) == 0 {
				releaseChanel <- 0
				return
			}

			time.Sleep(duration)
		}
	}
}

func (game *Game) start() chan int {
	if game.status() == STARTED {
		return game.ReleaseChanel
	}
	game.Status = STARTED
	game.seed.init(game.ticketConfig().MaxNumer)

	go game.seed.autoRelease(game.Interval, game.ReleaseChanel, game.QuitChanel)

	return game.ReleaseChanel
}

func (game *Game) stop() {
	if game.status() == STOPPED {
		return
	}
	game.Status = STOPPED
	game.QuitChanel <- true
	close(game.ReleaseChanel)
	close(game.QuitChanel)
}

func (game *Game) pause() {
	if game.status() == PAUSED {
		return
	}
	game.Status = PAUSED
	game.QuitChanel <- true
}

func (game *Game) resume() {
	if game.status() == STARTED {
		return
	}
	game.Status = STARTED
	go game.seed.autoRelease(game.Interval, game.ReleaseChanel, game.QuitChanel)
}

func (game *Game) addResultSeed(number int) {
	game.resultSeed.push(number)
}

func (game *Game) result() []int {
	return game.resultSeed.numbers
}

func (game *Game) isStarted() bool {
	return game.Status == STARTED
}

func (game *Game) isPaused() bool {
	return game.Status == PAUSED
}

func (game *Game) status() GameStatus {
	return game.Status
}

func (game *Game) ticketConfig() TicketConifg {
	return game.TicketConifg
}
