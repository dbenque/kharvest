package server

import (
	"github.com/dbenque/kharvest/store"
	"github.com/dbenque/kharvest/util"
)

var serverStore store.Store

func init() {
	serverStore = store.NewInMemStore(util.BuildKeyString, 30)
}

//GetStore return the store to be used by the servers
func GetStore() store.Store {
	return serverStore
}
