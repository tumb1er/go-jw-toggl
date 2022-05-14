package store

//go:generate bash -c "sed 's/Foo/Client/g;s/foo/gtclient/g' foo.tpl > clients_gen.go"

import (
	"go-jw-toggl/jwtoggl/gtoggl/gtclient"
	"gorm.io/gorm"
)

type Client struct {
	Synced
	gtclient.Client
}

type Clients []Client

type ClientManager gorm.DB

func (m *ClientManager) List() Clients {
	return listClients(m.gorm())
}

func (qs *Clients) Preload(_ *gorm.DB) error {
	return nil
}
