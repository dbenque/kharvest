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
	//Return true if the store already has at least one reference of such a {file,version} (ie with same content). First bool, file is known; Second bool, version already known
	Has(signature *pb.DataSignature) (bool, bool)
	//Store the content of the file. Return the number of version and the number of reference for the given version.
	Store(signature *pb.Data) Action
}
