package server

import (
	"encoding/base64"
	"log"
	"net"
	"os"
	"sync"

	pb "github.com/dbenque/kharvest/kharvest"
	"github.com/dbenque/kharvest/store"
	"github.com/dbenque/kharvest/util"
	"golang.org/x/net/context"

	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":80"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	storage  store.Store
	dedupers DeduperMap
}

var _ pb.KharvestServer = &server{}

//RunKharvestServer run the server for kharvest
func RunKharvestServer(storage store.Store) {
	log.Println("[kharvest] starting server...")
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("[kharvest] [error] failed to listen: %v", err)
		os.Exit(1)
	}
	grpcServer := grpc.NewServer()
	server := &server{
		storage:  storage,
		dedupers: DeduperMap{m: map[string]*Deduper{}},
	}
	server.initDedupers()
	pb.RegisterKharvestServer(grpcServer, server)
	// Register reflection service on gRPC server.
	reflection.Register(grpcServer)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("[kharvest] [error] failed to serve: %v", err)
		os.Exit(1)
	}
}

//Deduper hold chan for a given sign
type Deduper struct {
	sync.Mutex
	storeRequestChan chan bool
	stopChan         chan struct{}
	stopped          bool
}

//DeduperMap map of deduper per key
type DeduperMap struct {
	sync.RWMutex
	m map[string]*Deduper
}

//BuildResponse wait for the deduper signal a generate associated Notification reply
func (d *Deduper) BuildResponse(key string) (*pb.NotifyReply, error) {
	for {
		select {
		case <-d.stopChan:
			return &pb.NotifyReply{Action: pb.NotifyReply_ACK}, nil
		case nextAttempt, ok := <-d.storeRequestChan:
			if !ok {
				return &pb.NotifyReply{Action: pb.NotifyReply_ACK}, nil
			}
			if nextAttempt {
				return &pb.NotifyReply{Action: pb.NotifyReply_STORE_REQUESTED}, nil
			}
		}
	}
}
func (d *Deduper) stop() {
	d.Lock()
	defer d.Unlock()
	if !d.stopped {
		d.stopped = true
		close(d.stopChan)
	}
}
func (d *Deduper) start(retryPeriod time.Duration) {
	go func() {
		tick := time.NewTicker(retryPeriod)
		defer tick.Stop()
		for {
			select {
			case <-d.stopChan:
				return
			case <-tick.C:
				d.storeRequestChan <- true // try another one
			}
		}
	}()
}

//initDedupers get all the keys from the store
func (s *server) initDedupers() {
	for k := range s.storage.GetKeys() {
		s.dedupers.m[k] = nil
	}
}

func (s *server) Notify(ctx context.Context, dataSignature *pb.DataSignature) (*pb.NotifyReply, error) {
	str64 := base64.StdEncoding.EncodeToString([]byte(dataSignature.GetMd5()))
	log.Printf("[kharvest] [%s/%s] [%s] [%s] Receive Notification", dataSignature.GetNamespace(), dataSignature.GetPodName(), dataSignature.GetFilename(), str64)
	key := util.BuildKeyString(dataSignature)
	//Try in read only mode
	s.dedupers.RLock()
	deduper, ok := s.dedupers.m[key]
	s.dedupers.RUnlock()
	if ok {
		if deduper == nil {
			log.Printf("[kharvest] [warning] [%s/%s] [%s] [%s] No deduper(1) -> Ack", dataSignature.GetNamespace(), dataSignature.GetPodName(), dataSignature.GetFilename(), str64)
			s.storage.Reference(dataSignature)
			return &pb.NotifyReply{Action: pb.NotifyReply_ACK}, nil
		}
		r, err := deduper.BuildResponse(key)
		if err == nil && r != nil && r.GetAction() == pb.NotifyReply_ACK {
			s.storage.Reference(dataSignature)
		}
		return r, err
	}

	//Ok grant write access
	s.dedupers.Lock()
	if deduper, ok := s.dedupers.m[key]; ok {
		s.dedupers.Unlock()
		if deduper == nil {
			log.Printf("[kharvest] [warning] [%s/%s] [%s] [%s] No deduper(2) -> Ack", dataSignature.GetNamespace(), dataSignature.GetPodName(), dataSignature.GetFilename(), str64)
			s.storage.Reference(dataSignature)
			return &pb.NotifyReply{Action: pb.NotifyReply_ACK}, nil
		}

		r, err := deduper.BuildResponse(key)
		if err == nil && r != nil && r.GetAction() == pb.NotifyReply_ACK {
			s.storage.Reference(dataSignature)
		}
		return r, err
	}

	deduper = &Deduper{storeRequestChan: make(chan bool), stopChan: make(chan struct{})}
	s.dedupers.m[key] = deduper
	s.dedupers.Unlock()
	deduper.start(2 * time.Second)
	log.Printf("[kharvest] [%s/%s] [%s] [%s] Send store request", dataSignature.GetNamespace(), dataSignature.GetPodName(), dataSignature.GetFilename(), str64)
	return &pb.NotifyReply{Action: pb.NotifyReply_STORE_REQUESTED}, nil
}

func (s *server) Store(ctx context.Context, data *pb.Data) (*pb.StoreReply, error) {
	str64 := base64.StdEncoding.EncodeToString([]byte(data.Signature.GetMd5()))
	log.Printf("[kharvest] [%s/%s] [%s] [%s] Store receive, length %d", data.Signature.GetNamespace(), data.Signature.GetPodName(), data.Signature.GetFilename(), str64, len(data.Data))

	key := util.BuildKeyString(data.Signature)
	s.dedupers.RLock()
	deduper, ok := s.dedupers.m[key]
	s.dedupers.RUnlock()
	if !ok {
		log.Printf("[kharvest] [%s/%s] [%s] [%s] Deduper was already cleaned.", data.Signature.GetNamespace(), data.Signature.GetPodName(), data.Signature.GetFilename(), str64)
		return &pb.StoreReply{}, nil
	}
	err := s.storage.Store(data)
	deduper.stop()

	if err != nil {
		log.Printf("[kharvest] [error] [%s/%s] [%s] [%s] Storage error: %v", data.Signature.GetNamespace(), data.Signature.GetPodName(), data.Signature.GetFilename(), str64, err)
		return nil, err
	}

	log.Printf("[kharvest] [%s/%s] [%s] [%s] Store complete", data.Signature.GetNamespace(), data.Signature.GetPodName(), data.Signature.GetFilename(), str64)

	return &pb.StoreReply{}, nil
}
