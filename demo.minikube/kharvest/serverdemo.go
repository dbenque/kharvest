package main

import (
	"github.com/dbenque/kharvest/server"
	"github.com/dbenque/kharvest/store"
	"github.com/dbenque/kharvest/util"
)

func main() {
	storage := store.NewInMemStore(util.BuildKeyString, 30)
	go server.RunKharvestServerUserAPI(storage)
	server.RunKharvestServer(storage)
}
