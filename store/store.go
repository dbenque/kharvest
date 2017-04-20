package store

import pb "github.com/dbenque/kharvest/kharvest"

//Action type of action performed by the store
type Action string

//Definition of Action enum
const (
	Create       Action = "created"      // First time this file was seen by the store
	NewVersion   Action = "newVersion"   // A new version of the file was created
	NewReference Action = "newReference" // A new reference was added to that {file,version}
	None         Action = "none"         // This {file,version} was already stored for same reference.
	Error        Action = "error"
)

//Store interface to persist and retrieve data of kharvest
type Store interface {
	//GetKeys return the key strings generated from the content of the store
	GetKeys(func(dataSignature *pb.DataSignature) string) map[string]struct{}
	//Store the content of the file. Return the number of version and the number of reference for the given version.
	Store(data *pb.Data) Action
	//Reference associate all the metadata of the signature to the underlying data that must have been previously stored
	Reference(dataSignature *pb.DataSignature) error
}
