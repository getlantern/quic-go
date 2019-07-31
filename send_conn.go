package quic

import (
	"net"
)

type envelope interface {
	EnvelopeSize() uint64
}

// A sendConn allows sending using a simple Write() on a non-connected packet conn.
type sendConn interface {
	Write([]byte) error
	Close() error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	EnvelopeSize() uint64
}

type conn struct {
	net.PacketConn

	remoteAddr net.Addr
}

var _ sendConn = &conn{}

func newSendConn(c net.PacketConn, remote net.Addr) sendConn {
	return &conn{PacketConn: c, remoteAddr: remote}
}

func (c *conn) Write(p []byte) error {
	_, err := c.PacketConn.WriteTo(p, c.remoteAddr)
	return err
}

func (c *conn) RemoteAddr() net.Addr {
	return c.remoteAddr
}

// EnvelopeSize returns the number of bytes that should be reserved
// in each datagram on this connection for envelope usage which influences
// the maximum size of QUIC packets generated.  Defaults
// to 0.  If the underlying net.PacketConn provides an EnvelopeSize() method,
// the result of that method is used.
func (c *conn) EnvelopeSize() uint64 {
	if ec, ok := c.PacketConn.(envelope); ok {
		return ec.EnvelopeSize()
	} else {
		return 0
	}
}
