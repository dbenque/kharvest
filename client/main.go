package main

import (
	"flag"
	"fmt"

	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dbenque/kharvest/fswatcher"
)

const (
	address     = "localhost:8989"
	defaultName = "world"
)

func main() {
	conf := NewConfig()
	flag.Var(&conf.files, "files", "List of files")
	flag.StringVar(&conf.configPath, "config", configPath, "Config with list of files")
	flag.Parse()
	if err := conf.ReadAndWatch(); err != nil {
		fmt.Fprintf(os.Stderr, "Can't read configuration: %v", err)
		return
	}

	fileWatcher := fswatcher.StartFileWatcher([]string{})

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	exitChan := make(chan struct{})

	//Termination of program
	go func() {
		<-sigs
		conf.StopWatching()
		time.Sleep(2 * time.Second) // graceful stop of watchers.
		close(exitChan)
	}()

	//Watching configuration changes
	go func() {
		currentConfig := []string{}
		for {
			files, ok := <-conf.filesChan
			if !ok {
				fmt.Printf("The configuration is now immutable: %v", currentConfig)
				break
			}
			currentConfig = files
			fileWatcher.Set(files)
			fmt.Printf("files: %#v\n", files)
		}
	}()

	// reacting to fileWatcher events.
	go func() {
		for {
			select {
			case event, ok := <-fileWatcher.GetEventChan():
				if !ok {
					return
				}
				fmt.Printf("The content of watched file %s, has changed:\n%s\n", event.Filepath, string(event.Content))
			}
		}
	}()

	<-exitChan
	// var wg sync.WaitGroup
	// for range []int{1, 2, 3} {
	// 	wg.Add(1)
	// 	go func() {
	// 		defer wg.Done()
	// 		// Set up a connection to the server.
	// 		conn, err := grpc.Dial(address, grpc.WithInsecure())
	// 		if err != nil {
	// 			log.Fatalf("did not connect: %v", err)
	// 		}
	// 		defer conn.Close()
	// 		c := pb.NewKharvestClient(conn)

	// 		data := pb.Data{
	// 			Data:      []byte("this is content for kharwest"),
	// 			Signature: &pb.DataSignature{Filename: "toto", Md5: "6"},
	// 		}
	// 		r, err := c.Notify(context.Background(), data.Signature)
	// 		if err != nil {
	// 			fmt.Printf("ERROR:%v\n", err)
	// 			os.Exit(1)
	// 		} else if r.Action == pb.NotifyReply_ACK {
	// 			fmt.Printf("ACK\n")
	// 		} else if r.Action == pb.NotifyReply_STORE_REQUESTED {
	// 			fmt.Printf("STORE REQUESTED\n")
	// 			if _, err := c.Store(context.Background(), &data); err != nil {
	// 				fmt.Printf("STORE ERROR:%v\n", err)
	// 			} else {
	// 				fmt.Printf("STORE COMPLETED\n")
	// 			}
	// 		}
	// 	}()
	// }
	// wg.Wait()
}
