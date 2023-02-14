package pkg

import (
	"testing"

	"github.com/apex/log"
	"github.com/stretchr/testify/require"
)

func Test_Init_Seeds(t *testing.T) {
	ticket := NewTicket(TicketConifg{
		MaxCol:         9,
		MaxRow:         9,
		MaxNumberOfRow: 5,
	})
	ticket.generateNumbers()
	log.Info(BeautyTicket(ticket.board))

	require.Equal(t, true, true)
}
