package quic

import (
	"net"
	"sync"
)

type envelope interface {
	EnvelopeSize() uint64
}

type connection interface {
	Write([]byte) error
	Read([]byte) (int, net.Addr, error)
	Close() error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	SetCurrentRemoteAddr(net.Addr)
	EnvelopeSize() uint64
}

type conn struct {
	mutex sync.RWMutex

	pconn       net.PacketConn
	currentAddr net.Addr
}

var _ connection = &conn{}

func (c *conn) Write(p []byte) error {
	_, err := c.pconn.WriteTo(p, c.currentAddr)
	return err
}

func (c *conn) Read(p []byte) (int, net.Addr, error) {
	return c.pconn.ReadFrom(p)
}

func (c *conn) SetCurrentRemoteAddr(addr net.Addr) {
	c.mutex.Lock()
	c.currentAddr = addr
	c.mutex.Unlock()
}

func (c *conn) LocalAddr() net.Addr {
	return c.pconn.LocalAddr()
}

func (c *conn) RemoteAddr() net.Addr {
	c.mutex.RLock()
	addr := c.currentAddr
	c.mutex.RUnlock()
	return addr
}

func (c *conn) Close() error {
	return c.pconn.Close()
}

// EnvelopeSize returns the number of bytes that should be reserved
// in each datagram on this connection for envelope usage which influences
// the maximum size of QUIC packets generated.  Defaults
// to 0.  If the underlying net.PacketConn provides an EnvelopeSize() method,
// the result of that method is used.
func (c *conn) EnvelopeSize() uint64 {
	if ec, ok := c.pconn.(envelope); ok {
		return ec.EnvelopeSize()
	} else {
		return 0
	}
}
