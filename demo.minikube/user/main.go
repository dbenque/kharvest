package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"google.golang.org/grpc"

	pb "github.com/dbenque/kharvest/kharvest"
	google_protobuf "github.com/golang/protobuf/ptypes/empty"
)

const (
	serviceName = "kharvest"
	servicePort = "80"
)

func main() {
	kharvestURL := flag.String("k", "kharvest:80", "URL to kharvest server")
	flag.Parse()
	conn, err := grpc.Dial(*kharvestURL, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("[kharvest] [error] grpc client error: %v", err) //Error
	}
	defer conn.Close()
	kharvestUserAPI := pb.NewKharvestUserAPIClient(conn)

	reply, err := kharvestUserAPI.Keys(context.Background(), &google_protobuf.Empty{})
	if err != nil {
		os.Stderr.Write([]byte(err.Error()))
		os.Exit(1)
	}
	fmt.Printf("%#v\n", reply)
}
