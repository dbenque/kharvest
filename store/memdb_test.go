package store

import (
	"testing"

	pb "github.com/dbenque/kharvest/kharvest"
	"github.com/dbenque/kharvest/util"
)

func TestMemdb(t *testing.T) {
	d1 := pb.Data{Signature: &pb.DataSignature{Filename: "testfilename", Md5: "5"}}
	d2 := pb.Data{Signature: &pb.DataSignature{Filename: "testfilename", Md5: "6"}}
	d3 := pb.Data{Signature: &pb.DataSignature{Filename: "othertestfilename", Md5: "6"}}

	inMemStoreInstance.Store(&d1)
	inMemStoreInstance.Store(&d2)
	inMemStoreInstance.Store(&d3)

	keys := inMemStoreInstance.GetKeys(util.BuildKeyString)
	if len(keys) != 3 {
		t.Errorf("Bad count %#v\n", keys)
	}
}
