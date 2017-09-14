package store

import (
	"encoding/base64"
	"fmt"
	"log"
	"strconv"
	"time"

	pb "github.com/dbenque/kharvest/kharvest"
	"github.com/dbenque/kharvest/util"
	"github.com/golang/protobuf/ptypes/timestamp"
	memdb "github.com/hashicorp/go-memdb"
)

func init() {
	inMemStoreInstance = NewInMemStore(util.BuildKeyString, 60)
}

//KeyBuilderFunction func that return a unique key for a given data
type KeyBuilderFunction func(dataSignature *pb.DataSignature) string

var inMemStoreInstance *InMemStore
var _ Store = &InMemStore{}

//InMemStore in memory store based on go-memdb
type InMemStore struct {
	db                     *memdb.MemDB
	keyBuilder             func(dataSignature *pb.DataSignature) string
	timePeriodIndexSeconds int64
}

//NewInMemStore create a new in memory storage
func NewInMemStore(keyBuilderFunc KeyBuilderFunction, timePeriodIndexSeconds int64) *InMemStore {

	if timePeriodIndexSeconds < 1 {
		timePeriodIndexSeconds = 1
	}
	inMemStore := &InMemStore{keyBuilder: keyBuilderFunc, timePeriodIndexSeconds: timePeriodIndexSeconds}
	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"data": &memdb.TableSchema{
				Name: "data",
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:    "id",
						Unique:  true,
						Indexer: &idDataIndexer{keyBuilderFunc},
					},
					"filename": &memdb.IndexSchema{
						Name:    "filename",
						Unique:  false,
						Indexer: &filenameIndexer{},
					},
				},
			},
			"reference": &memdb.TableSchema{
				Name: "reference",
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:   "id",
						Unique: true,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "Filename"},
								&memdb.StringFieldIndex{Field: "Namespace"},
								&memdb.StringFieldIndex{Field: "PodName"},
								&timestampIndexer{},
							},
						},
					},
					"pod": &memdb.IndexSchema{
						Name:   "pod",
						Unique: false,
						Indexer: &memdb.CompoundIndex{
							Indexes: []memdb.Indexer{
								&memdb.StringFieldIndex{Field: "Namespace"},
								&memdb.StringFieldIndex{Field: "PodName"},
							},
						},
					},
					"at": &memdb.IndexSchema{
						Name:    "at",
						Unique:  false,
						Indexer: &timestampIndexer{timePeriodIndexSeconds},
					},
					"datakey": &memdb.IndexSchema{
						Name:    "datakey",
						Unique:  false,
						Indexer: &idDataIndexer{keyBuilderFunc},
					},
					"meta": &memdb.IndexSchema{
						Name:    "meta",
						Unique:  false,
						Indexer: &memdb.StringMapFieldIndex{Field: "Metadata", Lowercase: false},
					},
				},
			},
		},
	}
	var err error
	inMemStore.db, err = memdb.NewMemDB(schema)
	if err != nil {
		panic(err)
	}
	return inMemStore
}

//Store the content of the file. Return the number of version and the number of reference for the given version.
func (ims *InMemStore) Store(data *pb.Data) error {
	if err := util.DefaultValueForSignature(data.GetSignature()); err != nil {
		return err
	}

	str64 := base64.StdEncoding.EncodeToString([]byte(data.Signature.GetMd5()))
	log.Printf("[kharvest] [InMemStore] [Store] %s/%s/%s MD5:%s", data.Signature.GetNamespace(), data.Signature.GetPodName(), data.Signature.GetFilename(), str64)

	txn := ims.db.Txn(true)
	if err := txn.Insert("data", data); err != nil {
		txn.Abort()
		return err
	}
	txn.Commit()

	return ims.Reference(data.GetSignature())
}

//Reference associate all the metadata of the signature to the underlying data that must have been previously stored
func (ims *InMemStore) Reference(dataSignature *pb.DataSignature) error {
	if err := util.DefaultValueForSignature(dataSignature); err != nil {
		return err
	}

	str64 := base64.StdEncoding.EncodeToString([]byte(dataSignature.GetMd5()))
	log.Printf("[kharvest] [InMemStore] [Reference] %s/%s/%s MD5:%s", dataSignature.GetNamespace(), dataSignature.GetPodName(), dataSignature.GetFilename(), str64)

	if dataSignature.Timestamp == nil {
		dataSignature.Timestamp = &timestamp.Timestamp{Seconds: time.Now().Unix()}
	}
	txn := ims.db.Txn(true)
	if err := txn.Insert("reference", dataSignature); err != nil {
		fmt.Println("Transaction abort:" + err.Error())
		txn.Abort()
		return err
	}
	txn.Commit()
	return nil
}

//GetKeys return the key strings generated from the content of the store
func (ims *InMemStore) GetKeys() map[string]struct{} {
	result := map[string]struct{}{}
	txn := ims.db.Txn(false)
	defer txn.Abort()
	iter, err := txn.Get("data", "id")
	if err != nil {
		return result
	}
	for v := iter.Next(); v != nil; v = iter.Next() {
		D := v.(*pb.Data)
		result[ims.keyBuilder(D.Signature)] = struct{}{}
	}
	return result
}

//GetSameReferences return all the references to the same data
func (ims *InMemStore) GetSameReferences(dataSignature *pb.DataSignature) []*pb.DataSignature {
	result := []*pb.DataSignature{}
	txn := ims.db.Txn(false)
	defer txn.Abort()
	iter, err := txn.Get("reference", "datakey", ims.keyBuilder(dataSignature))
	if err != nil {
		return result
	}
	for v := iter.Next(); v != nil; v = iter.Next() {
		s := v.(*pb.DataSignature)
		result = append(result, s)
	}
	return result
}

//GetPodReferences return all the references for a given pod
func (ims *InMemStore) GetPodReferences(namespace, podName string) []*pb.DataSignature {
	result := []*pb.DataSignature{}
	txn := ims.db.Txn(false)
	defer txn.Abort()
	iter, err := txn.Get("reference", "pod", namespace, podName)
	if err != nil {
		return result
	}
	for v := iter.Next(); v != nil; v = iter.Next() {
		s := v.(*pb.DataSignature)
		result = append(result, s)
	}
	return result
}

func (ims *InMemStore) getRefAt(at time.Time) []*pb.DataSignature {
	result := []*pb.DataSignature{}
	tSec := at.Unix()
	if ims.timePeriodIndexSeconds > 1 {
		tSec = tSec - tSec%ims.timePeriodIndexSeconds
	}

	txn := ims.db.Txn(false)
	defer txn.Abort()
	iter, err := txn.Get("reference", "at", strconv.FormatInt(tSec, 36))
	if err != nil {
		return result
	}
	for v := iter.Next(); v != nil; v = iter.Next() {
		s := v.(*pb.DataSignature)
		result = append(result, s)
	}
	return result
}

//GetReferencesAt return all the references for a given pod
func (ims *InMemStore) GetReferencesAt(from, to time.Time) []*pb.DataSignature {
	result := []*pb.DataSignature{}
	if to.Unix() <= from.Unix() {
		return result
	}

	tSec := from
	for tSec.Unix() < to.Unix() {
		for _, v := range ims.getRefAt(tSec) {
			if v.Timestamp.Seconds >= from.Unix() && v.Timestamp.Seconds < to.Unix() {
				result = append(result, v)
			}
		}
		tSec = tSec.Add(time.Duration(ims.timePeriodIndexSeconds) * time.Second)
	}
	// Do it one more time to capture valid intersection on last time window
	for _, v := range ims.getRefAt(tSec) {
		if v.Timestamp.Seconds >= from.Unix() && v.Timestamp.Seconds < to.Unix() {
			result = append(result, v)
		}
	}
	return result
}

//GetReferencesForMeta return all the references for a given pod
func (ims *InMemStore) GetReferencesForMeta(key, value string) []*pb.DataSignature {
	result := []*pb.DataSignature{}
	txn := ims.db.Txn(false)
	defer txn.Abort()
	iter, err := txn.Get("reference", "meta", key, value)
	if err != nil {
		return result
	}
	for v := iter.Next(); v != nil; v = iter.Next() {
		s := v.(*pb.DataSignature)
		result = append(result, s)
	}
	return result
}
