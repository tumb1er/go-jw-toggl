package store

//go:generate bash -c "sed 's/Foo/Client/g;s/foo/gtclient/g' foo.tpl > clients_gen.go"

import (
	"github.com/jinzhu/gorm"
	"go-jw-toggl/jwtoggl/dto"
	"go-jw-toggl/jwtoggl/gtoggl/gtclient"
)

type Client struct {
	Synced
	gtclient.Client
}

type Clients []Client

type ClientManager gorm.DB

func (c Client) GetTogglID() uint64 {
	return c.Id
}

func (c Client) GetName() string {
	return c.Name
}

func (qs Clients) DTO() []dto.Client {
	result := make([]dto.Client, len(qs))
	for i, c := range qs {
		result[i] = c
	}
	return result
}

func (m *ClientManager) List() Clients {
	return listClients(m.gorm())
}
