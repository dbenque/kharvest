package store

import (
	"time"

	pb "github.com/dbenque/kharvest/kharvest"
)

//Store give access to kharvest data storage
type Store interface {
	Writer
	Reader
	//GetKeyBuilder() func(dataSignature *pb.DataSignature) string
}

//Writer interface to persist data of kharvest
type Writer interface {
	//Store the content of the file. Return the number of version and the number of reference for the given version.
	Store(data *pb.Data) error
	//Reference associate all the metadata of the signature to the underlying data that must have been previously stored
	Reference(dataSignature *pb.DataSignature) error
}

//Reader interface to consume data of kharvest
type Reader interface {
	//GetKeys return the key strings generated from the content of the store
	GetKeys() map[string]struct{}
	//GetSameReferences return all the references to the same data
	GetSameReferences(dataSignature *pb.DataSignature) []*pb.DataSignature
	//GetPodReferences return all the references for a given pod
	GetPodReferences(namespace, podName string) []*pb.DataSignature
	//GetReferencesAt return all the references for a given pod
	GetReferencesAt(from, to time.Time) []*pb.DataSignature
	//GetReferencesForMeta return all the references for a given pod
	GetReferencesForMeta(key, value string) []*pb.DataSignature
}
