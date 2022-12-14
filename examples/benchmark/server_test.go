package benchmark

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/starine/aim"
	"github.com/starine/aim/examples/mock"
	"github.com/starine/aim/logger"
	"github.com/starine/aim/websocket"
)

const wsurl = "ws://localhost:8000"

func Test_Parallel(t *testing.T) {
	const count = 10000
	gpool, _ := ants.NewPool(50, ants.WithPreAlloc(true))
	defer gpool.Release()
	var wg sync.WaitGroup
	wg.Add(count)

	clis := make([]aim.Client, count)
	t0 := time.Now()
	for i := 0; i < count; i++ {
		idx := i
		_ = gpool.Submit(func() {
			cli := websocket.NewClient(fmt.Sprintf("test_%v", idx), "client", websocket.ClientOptions{
				Heartbeat: aim.DefaultHeartbeat,
			})
			// set dialer
			cli.SetDialer(&mock.WebsocketDialer{})

			// step2: 建立连接
			err := cli.Connect(wsurl)
			if err != nil {
				logger.Error(err)
			}
			clis[idx] = cli
			wg.Done()
		})
	}
	wg.Wait()
	t.Logf("logined %d cost %v", count, time.Since(t0))
	t.Logf("done connecting")
	time.Sleep(time.Second * 5)
	t.Logf("closed")

	for i := 0; i < count; i++ {
		clis[i].Close()
	}
}

func Test_Message(t *testing.T) {
	const count = 1000 * 100
	cli := websocket.NewClient(fmt.Sprintf("test_%v", 1), "client", websocket.ClientOptions{
		Heartbeat: aim.DefaultHeartbeat,
	})
	// set dialer
	cli.SetDialer(&mock.WebsocketDialer{})

	// step2: 建立连接
	err := cli.Connect(wsurl)
	if err != nil {
		logger.Error(err)
	}
	msg := []byte(strings.Repeat("hello", 10))
	t0 := time.Now()
	go func() {
		for i := 0; i < count; i++ {
			_ = cli.Send(msg)
		}
	}()
	recv := 0
	for {
		frame, err := cli.Read()
		if err != nil {
			logger.Info("time", time.Now().UnixNano(), err)
			break
		}
		if frame.GetOpCode() != aim.OpBinary {
			continue
		}
		recv++
		if recv == count { // 接收完消息
			break
		}
	}

	t.Logf("message %d cost %v", count, time.Since(t0))
}
