package main

import (
	"log"
	"net"
	"sync"

	pb "github.com/dbenque/kharvest/kharvest"
	"github.com/dbenque/kharvest/store"
	"golang.org/x/net/context"

	"time"

	"fmt"

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

//Deduper hold chan for a given signature
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

var dedupers DeduperMap

func init() {
	dedupers = DeduperMap{m: map[string]*Deduper{}}
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
				fmt.Println("NextAttempt")
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
		fmt.Println("Start tick")
		tick := time.NewTicker(retryPeriod)
		defer tick.Stop()
		for {
			select {
			case <-d.stopChan:
				return
			case <-tick.C:
				fmt.Println("Tick")
				d.storeRequestChan <- true // try another one
			}
		}
	}()
}

func (s *server) Notify(ctx context.Context, dataSignature *pb.DataSignature) (*pb.NotifyReply, error) {
	key := dataSignature.Filename + "." + dataSignature.Md5
	//Try in read only mode
	dedupers.RLock()
	deduper, ok := dedupers.m[key]
	dedupers.RUnlock()
	if ok {
		return deduper.BuildResponse(key)
	}

	//Ok grant write access
	dedupers.Lock()
	if deduper, ok := dedupers.m[key]; ok {
		dedupers.Unlock()
		return deduper.BuildResponse(key)
	}

	deduper = &Deduper{storeRequestChan: make(chan bool), stopChan: make(chan struct{})}
	dedupers.m[key] = deduper
	dedupers.Unlock()
	deduper.start(2 * time.Second)
	return &pb.NotifyReply{Action: pb.NotifyReply_STORE_REQUESTED}, nil
}

func (s *server) Store(ctx context.Context, data *pb.Data) (*pb.StoreReply, error) {
	key := data.Signature.Filename + "." + data.Signature.Md5
	dedupers.RLock()
	deduper, ok := dedupers.m[key]
	dedupers.RUnlock()
	if !ok {
		fmt.Println("Deduper was already cleaned!")
		return &pb.StoreReply{}, nil
	}
	action := s.storage.Store(data)
	fmt.Printf("Store=%v for %s", action, key)
	deduper.stop()

	return &pb.StoreReply{}, nil
}
