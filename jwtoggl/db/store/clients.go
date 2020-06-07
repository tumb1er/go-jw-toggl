package store

//go:generate bash -c "sed 's/Foo/Client/g;s/foo/gtclient/g' foo.tpl > clients_gen.go"

import (
	"github.com/jinzhu/gorm"
	"go-jw-toggl/jwtoggl/gtoggl/gtclient"
)

type Client struct {
	Synced
	gtclient.Client `gorm:"-"`
}

type Clients []Client

type ClientManager gorm.DB

func (m *ClientManager) List() gtclient.Clients {
	return listClients(m.gorm())
}
