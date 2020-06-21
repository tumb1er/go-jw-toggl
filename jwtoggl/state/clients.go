package state

//go:generate bash -c "sed 's/Foo/Client/g;s/foo/gtclient/g' foo.tpl > clients_gen.go"
//go:generate bash -c "sed 's/Foo/Client/g;s/foo/gtclient/g' bar.tpl >> clients_gen.go"

import (
	"go-jw-toggl/jwtoggl/dto"
)

type OnClientsChangedCallback func(dto.Clients)

type ClientsDB struct {
	common
	onClientsChanged OnClientsChangedCallback
}
