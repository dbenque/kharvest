package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dbenque/kharvest/client"
	"github.com/dbenque/toKube/deployer"
)

func main() {
	log.Println("Starting kharvestclient")
	flag.Parse()
	deployer.AutoDeploy()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	exitChan := make(chan struct{})

	conf := client.NewConfig()
	conf.ConfigPath = "/cfg/kharvest-client/kharvest.cfg"
	flag.Parse()
	if err := conf.ReadAndWatch(); err != nil {
		fmt.Fprintf(os.Stderr, "Can't read configuration: %v", err)
		return
	}

	//Termination of program
	go func() {
		<-sigs
		conf.StopWatching()
		time.Sleep(2 * time.Second) // graceful stop of watchers.
		close(exitChan)
	}()
	log.Println("Running KharvestClient")
	client.RunKharvestClient(conf)
	<-exitChan
}
