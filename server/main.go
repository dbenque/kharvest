package main

import (
	"log"
	"net"

	pb "github.com/dbenque/kharvest/kharvest"
	"github.com/dbenque/kharvest/store"
	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":8989"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	storage store.Store
}

var _ pb.KharvestServer = &server{}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterKharvestServer(s, &server{
		storage: store.NewInMemoryStorage(),
	})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (s *server) Notify(ctx context.Context, dataSignature *pb.DataSignature) (*pb.NotifyReply, error) {
	knownFile, knownVersion := s.storage.Has(dataSignature)
	if knownFile && knownVersion {
		return &pb.NotifyReply{Action: pb.NotifyReply_ACK}, nil
	}
	return &pb.NotifyReply{Action: pb.NotifyReply_STORE_REQUESTED}, nil
}

func (s *server) Store(context.Context, *pb.Data) (*pb.StoreReply, error) {

	return nil, nil
}
