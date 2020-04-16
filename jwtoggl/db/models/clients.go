package models

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

func ListClients(db *gorm.DB) gtclient.Clients {
	qs := NewClients()
	length := Must(db.Find(qs))
	result := make(gtclient.Clients, length)
	for i, c := range *qs {
		result[i] = c.Client
	}
	return result
}
