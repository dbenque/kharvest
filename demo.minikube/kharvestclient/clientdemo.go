package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dbenque/kharvest/client"
)

func main() {
	log.Println("Starting kharvestclient")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	exitChan := make(chan struct{})

	conf := client.NewConfig(os.Getenv("PODNAME"), os.Getenv("NAMESPACE"))
	conf.ConfigPath = "/cfg/kharvest-client/kharvest.cfg"

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
