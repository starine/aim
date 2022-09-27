package websocket

import (
	"bufio"
	"net"

	"github.com/gobwas/ws"
	"github.com/starine/aim"
)

// Server is a websocket implement of the Server
type Upgrader struct {
}

// NewServer NewServer
func NewServer(listen string, service aim.ServiceRegistration, options ...aim.ServerOption) aim.Server {
	return aim.NewServer(listen, service, new(Upgrader), options...)
}

func (u *Upgrader) Name() string {
	return "websocket.Server"
}

func (u *Upgrader) Upgrade(rawconn net.Conn, rd *bufio.Reader, wr *bufio.Writer) (aim.Conn, error) {
	_, err := ws.Upgrade(rawconn)
	if err != nil {
		return nil, err
	}
	conn := NewConnWithRW(rawconn, rd, wr)
	return conn, nil
}
