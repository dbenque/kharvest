package server

import (
	"log"
	"net"
	"os"

	pb "github.com/dbenque/kharvest/kharvest"
	"github.com/dbenque/kharvest/store"
	google_protobuf "github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	portAPIUser = ":81"
)

// serverUserAPI is used to implement customer facing API of kharvest
type serverUserAPI struct {
	storage store.Store
}

var _ pb.KharvestUserAPIServer = &serverUserAPI{}

//RunKharvestServerUserAPI Runs the userAPI server for kharvest
func RunKharvestServerUserAPI() {
	log.Println("[kharvestuserAPI] starting server...")
	lis, err := net.Listen("tcp", portAPIUser)
	if err != nil {
		log.Fatalf("[kharvestuserAPI] [error] failed to listen: %v", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	server := &serverUserAPI{
		storage: GetStore(),
	}
	pb.RegisterKharvestUserAPIServer(grpcServer, server)
	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("[kharvestuserAPI] [error] failed to serve: %v", err)
		os.Exit(1)
	}
}

func (s *serverUserAPI) Keys(context.Context, *google_protobuf.Empty) (*pb.KeysReply, error) {
	log.Printf("[kharvestUserAPI] Keys")

	keys := s.storage.GetKeys()
	reply := &pb.KeysReply{Keys: make([]string, len(keys))}
	i := 0
	for key := range keys {
		reply.Keys[i] = key
		i++
	}
	return reply, nil
}
