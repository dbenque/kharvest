package client

import (
	"context"
	"log"
	"time"

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

//RunKharvestClient Run the kharvest client
func RunKharvestClient(conf *Config) {
	fileWatcher := fswatcher.StartFileWatcher([]string{})
	//Watching configuration changes
	go func() {
		currentConfig := []string{}
		for {
			files, ok := <-conf.filesChan
			if !ok {
				log.Printf("[kharvest] The configuration is now immutable: %v", currentConfig)
				break
			}
			currentConfig = files
			fileWatcher.Set(files)
			log.Printf("[kharvest] files: %#v\n", files)
		}
	}()

	// reacting to fileWatcher events.
	go func() {
		conn, err := grpc.Dial(net.JoinHostPort(serviceName, servicePort), grpc.WithInsecure())
		if err != nil {
			log.Fatalf("[kharvest] [error] grpc client error: %v", err) //Error
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
				Signature: &pb.DataSignature{Filename: event.Filepath, Md5: string(event.MD5[:md5.Size]), PodName: conf.podName, Namespace: conf.namespace},
			}
			ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
			r, err := kharvestServer.Notify(ctx, data.Signature)
			cancelFunc()
			if err != nil {
				log.Printf("[kharvest] [error] Error on Notify (item will be resent): %v", err) //Error
				fileWatcher.Resend(event)
			} else if r.Action == pb.NotifyReply_ACK {
				log.Printf("[kharvest] Server ack for file: %s", event.Filepath)
			} else if r.Action == pb.NotifyReply_STORE_REQUESTED {
				log.Printf("[kharvest] Server requests store for file: %s", event.Filepath)
				if _, err := kharvestServer.Store(context.Background(), &data); err != nil {
					log.Printf("[kharvest] [error]  Server fails storing file %s, error: %v", event.Filepath, err) //Error
				} else {
					log.Printf("[kharvest] Storage done for file: %s", event.Filepath)
				}
			}
		}
	}()
}
