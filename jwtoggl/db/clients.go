package db

//go:generate bash -c "sed 's/Foo/Client/g;s/foo/gtclient/g' foo.tpl > clients_gen.go"
//go:generate bash -c "sed 's/Foo/Client/g;s/foo/gtclient/g' bar.tpl >> clients_gen.go"

import (
	"go-jw-toggl/jwtoggl/gtoggl/gtclient"
)

type OnClientsChangedCallback func(gtclient.Clients)

type ClientsDB struct {
	CommonDB
	onClientsChanged OnClientsChangedCallback
}
