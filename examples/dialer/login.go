package dialer

import (
	"github.com/starine/aim"
	"github.com/starine/aim/websocket"
	"github.com/starine/aim/wire/token"
)

func Login(wsurl, account string, appSecrets ...string) (aim.Client, error) {
	cli := websocket.NewClient(account, "unittest", websocket.ClientOptions{})
	secret := token.DefaultSecret
	if len(appSecrets) > 0 {
		secret = appSecrets[0]
	}
	cli.SetDialer(&ClientDialer{
		AppSecret: secret,
	})
	err := cli.Connect(wsurl)
	if err != nil {
		return nil, err
	}
	return cli, nil
}
