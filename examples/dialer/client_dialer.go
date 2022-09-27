package dialer

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/starine/aim"
	"github.com/starine/aim/logger"
	"github.com/starine/aim/wire"
	"github.com/starine/aim/wire/pkt"
	"github.com/starine/aim/wire/token"
)

type ClientDialer struct {
	AppSecret string
}

func (d *ClientDialer) DialAndHandshake(ctx aim.DialerContext) (net.Conn, error) {
	// 1. 拨号
	conn, _, _, err := ws.Dial(context.TODO(), ctx.Address)
	if err != nil {
		return nil, err
	}
	if d.AppSecret == "" {
		d.AppSecret = token.DefaultSecret
	}
	// 2. 直接使用封装的JWT包生成一个token
	tk, err := token.Generate(d.AppSecret, &token.Token{
		Account: ctx.Id,
		App:     "aim",
		Exp:     time.Now().AddDate(0, 0, 1).Unix(),
	})
	if err != nil {
		return nil, err
	}
	// 3. 发送一条CommandLoginSignIn消息
	loginreq := pkt.New(wire.CommandLoginSignIn).WriteBody(&pkt.LoginReq{
		Token: tk,
	})
	err = wsutil.WriteClientBinary(conn, pkt.Marshal(loginreq))
	if err != nil {
		return nil, err
	}

	// wait resp
	_ = conn.SetReadDeadline(time.Now().Add(ctx.Timeout))
	frame, err := ws.ReadFrame(conn)
	if err != nil {
		return nil, err
	}
	ack, err := pkt.MustReadLogicPkt(bytes.NewBuffer(frame.Payload))
	if err != nil {
		return nil, err
	}
	// 4. 判断是否登录成功
	if ack.Status != pkt.Status_Success {
		return nil, fmt.Errorf("login failed: %v", &ack.Header)
	}
	var resp = new(pkt.LoginResp)
	_ = ack.ReadBody(resp)

	logger.Debug("logined ", resp.GetChannelId())
	return conn, nil
}
