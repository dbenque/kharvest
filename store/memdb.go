package store

import (
	"fmt"

	pb "github.com/dbenque/kharvest/kharvest"
	"github.com/dbenque/kharvest/util"
	memdb "github.com/hashicorp/go-memdb"
)

type idIndexer struct {
}

func (s *idIndexer) FromObject(obj interface{}) (bool, []byte, error) {
	data, ok := obj.(*pb.Data)
	if !ok {
		return false, nil, fmt.Errorf("Unknown object type. Not a pb.Data")
	}
	val := util.BuildKeyString(data.Signature)
	val += "\x00"
	return true, []byte(val), nil
}

func (s *idIndexer) FromArgs(args ...interface{}) ([]byte, error) {
	return fromArgs(args...)
}

type filenameIndexer struct {
}

func (s *filenameIndexer) FromObject(obj interface{}) (bool, []byte, error) {
	data, ok := obj.(*pb.Data)
	if !ok {
		return false, nil, fmt.Errorf("Unknown object type. Not a pb.Data")
	}
	val := data.Signature.Filename
	val += "\x00"
	return true, []byte(val), nil
}

func (s *filenameIndexer) FromArgs(args ...interface{}) ([]byte, error) {
	return fromArgs(args...)
}

func fromArgs(args ...interface{}) ([]byte, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("must provide only a single argument")
	}
	arg, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("argument must be a string: %#v", args[0])
	}
	// Add the null character as a terminator
	arg += "\x00"
	return []byte(arg), nil
}

var inMemStoreInstance InMemStore

func init() {
	inMemStoreInstance.schema = &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"data": &memdb.TableSchema{
				Name: "data",
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:    "id",
						Unique:  true,
						Indexer: &idIndexer{},
					},
					"filename": &memdb.IndexSchema{
						Name:    "filename",
						Unique:  false,
						Indexer: &filenameIndexer{},
					},
				},
			},
		},
	}
	var err error
	inMemStoreInstance.db, err = memdb.NewMemDB(inMemStoreInstance.schema)
	if err != nil {
		panic(err)
	}
}

//InMemStore in memory store based on go-memdb
type InMemStore struct {
	schema *memdb.DBSchema
	db     *memdb.MemDB
}

//GetKeys return the key strings generated from the content of the store
func (ims *InMemStore) GetKeys(func(dataSignature *pb.DataSignature) string) map[string]struct{} {
	result := map[string]struct{}{}
	txn := ims.db.Txn(false)
	defer txn.Abort()
	iter, err := txn.Get("data", "id")
	if err != nil {
		return result
	}
	for v := iter.Next(); v != nil; v = iter.Next() {
		D := v.(*pb.Data)
		result[util.BuildKeyString(D.Signature)] = struct{}{}
	}
	return result
}

//Store the content of the file. Return the number of version and the number of reference for the given version.
func (ims *InMemStore) Store(data *pb.Data) Action {
	txn := ims.db.Txn(true)
	if err := txn.Insert("data", data); err != nil {
		fmt.Println(err.Error())
		txn.Abort()
		return Error
	}
	txn.Commit()
	return Create
}

//Reference associate all the metadata of the signature to the underlying data that must have been previously stored
func (ims *InMemStore) Reference(dataSignature *pb.DataSignature) error {
	return nil
}
