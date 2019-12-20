package quic

import (
	"errors"
	"testing"

	"github.com/lucas-clemente/quic-go/internal/utils"
)

func TestPacketHandlerMapDeadlock(t *testing.T) {
	// conn that immediately returns an error on read
	conn := newMockPacketConn()
	conn.readErr = errors.New("uh oh")

	// manually construct to prevent immediate launch of listen goroutine
	handler := &packetHandlerMap{
		conn:      conn,
		listening: make(chan struct{}),
		logger:    utils.DefaultLogger,
	}

	server := &server{
		sessionHandler: handler,
		sessionQueue:   make(chan Session),
		errorChan:      make(chan struct{}),
		logger:         utils.DefaultLogger,
	}

	handler.SetServer(server)

	go handler.listen()

	// this should immediately error.
	_, err := server.Accept()
	if err == nil {
		t.Errorf("expected non-nil error...")
	}

}
