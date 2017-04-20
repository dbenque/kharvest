package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"os"

	pb "github.com/dbenque/kharvest/kharvest"
	"google.golang.org/grpc"
)

const (
	address     = "localhost:8989"
	defaultName = "world"
)

func main() {
	var wg sync.WaitGroup
	for range []int{1, 2, 3} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Set up a connection to the server.
			conn, err := grpc.Dial(address, grpc.WithInsecure())
			if err != nil {
				log.Fatalf("did not connect: %v", err)
			}
			defer conn.Close()
			c := pb.NewKharvestClient(conn)

			data := pb.Data{
				Data:      []byte("this is content for kharwest"),
				Signature: &pb.DataSignature{Filename: "toto", Md5: "6"},
			}
			r, err := c.Notify(context.Background(), data.Signature)
			if err != nil {
				fmt.Printf("ERROR:%v\n", err)
				os.Exit(1)
			} else if r.Action == pb.NotifyReply_ACK {
				fmt.Printf("ACK\n")
			} else if r.Action == pb.NotifyReply_STORE_REQUESTED {
				fmt.Printf("STORE REQUESTED\n")
				if _, err := c.Store(context.Background(), &data); err != nil {
					fmt.Printf("STORE ERROR:%v\n", err)
				} else {
					fmt.Printf("STORE COMPLETED\n")
				}
			}
		}()
	}
	wg.Wait()
}
