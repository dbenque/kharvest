package client

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"

	"crypto/md5"

	"net"

	"github.com/dbenque/kharvest/fswatcher"
	pb "github.com/dbenque/kharvest/kharvest"
)

const (
	serviceName = "kharvest"
	servicePort = "80"
)

// func main() {
// 	sigs := make(chan os.Signal, 1)
// 	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
// 	exitChan := make(chan struct{})

// 	conf := NewConfig()
// 	flag.Var(&conf.files, "files", "List of files")
// 	flag.StringVar(&conf.configPath, "config", configPath, "Config with list of files")
// 	flag.Parse()
// 	if err := conf.ReadAndWatch(); err != nil {
// 		fmt.Fprintf(os.Stderr, "Can't read configuration: %v", err)
// 		return
// 	}

// 	//Termination of program
// 	go func() {
// 		<-sigs
// 		conf.StopWatching()
// 		time.Sleep(2 * time.Second) // graceful stop of watchers.
// 		close(exitChan)
// 	}()
// 	runKharvestClient(conf)
// 	<-exitChan
// }

//RunKharvestClient Run the kharvest client
func RunKharvestClient(conf *Config) {
	fileWatcher := fswatcher.StartFileWatcher([]string{})
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
		conn, err := grpc.Dial(net.JoinHostPort(serviceName, servicePort), grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		kharvestServer := pb.NewKharvestClient(conn)

		for {
			event, ok := <-fileWatcher.GetEventChan()
			if !ok {
				return
			}
			data := pb.Data{
				Data:      event.Content,
				Signature: &pb.DataSignature{Filename: event.Filepath, Md5: string(event.MD5[:md5.Size])},
			}
			r, err := kharvestServer.Notify(context.Background(), data.Signature)
			if err != nil {
				fmt.Printf("ERROR:%v\n", err)
			} else if r.Action == pb.NotifyReply_ACK {
				fmt.Printf("ACK\n")
			} else if r.Action == pb.NotifyReply_STORE_REQUESTED {
				fmt.Printf("STORE REQUESTED\n")
				if _, err := kharvestServer.Store(context.Background(), &data); err != nil {
					fmt.Printf("STORE ERROR:%v\n", err)
				} else {
					fmt.Printf("STORE COMPLETED\n")
				}
			}
		}
	}()

}
