package main

import (
	"flag"

	"github.com/dbenque/kharvest/server"
	"github.com/dbenque/toKube/deployer"
)

func main() {
	flag.Parse()
	deployer.AutoDeploy()

	go server.RunKharvestServerUserAPI()
	server.RunKharvestServer()

}
