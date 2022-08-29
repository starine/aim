package serv

import (
	"net"

	"github.com/starine/aim"
	"github.com/starine/aim/logger"
	"github.com/starine/aim/tcp"
	"github.com/starine/aim/wire/pkt"
	"google.golang.org/protobuf/proto"
)

type TcpDialer struct {
	ServiceId string
}

func NewDialer(serviceId string) aim.Dialer {
	return &TcpDialer{
		ServiceId: serviceId,
	}
}

// DialAndHandshake(context.Context, string) (net.Conn, error)
func (d *TcpDialer) DialAndHandshake(ctx aim.DialerContext) (net.Conn, error) {
	// 1. 拨号建立连接
	conn, err := net.DialTimeout("tcp", ctx.Address, ctx.Timeout)
	if err != nil {
		return nil, err
	}
	req := &pkt.InnerHandshakeReq{
		ServiceId: d.ServiceId,
	}
	logger.Infof("send req %v", req)
	// 2. 把自己的ServiceId发送给对方
	bts, _ := proto.Marshal(req)
	err = tcp.WriteFrame(conn, aim.OpBinary, bts)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
