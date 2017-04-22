package store

import (
	"testing"

	"time"

	pb "github.com/dbenque/kharvest/kharvest"
)

func TestMemdbStoreData(t *testing.T) {
	d1 := pb.Data{Signature: &pb.DataSignature{Filename: "testfilename", Md5: "5"}}
	d2 := pb.Data{Signature: &pb.DataSignature{Filename: "testfilename", Md5: "6"}}
	d3 := pb.Data{Signature: &pb.DataSignature{Filename: "othertestfilename", Md5: "6"}}

	inMemStoreInstance.Store(&d1)
	inMemStoreInstance.Store(&d2)
	inMemStoreInstance.Store(&d3)

	keys := inMemStoreInstance.GetKeys()
	if len(keys) != 3 {
		t.Errorf("Bad count %#v\n", keys)
	}
}

func TestMemdbStoreDataSignature(t *testing.T) {
	d1 := pb.Data{Signature: &pb.DataSignature{Filename: "testfilename", Md5: "5", Namespace: "toto", PodName: "p1", Metadata: map[string]string{"app": "AAA", "phase": "PDT"}}}
	d11 := pb.Data{Signature: &pb.DataSignature{Filename: "testfilename", Md5: "5", Namespace: "toto", PodName: "p11", Metadata: map[string]string{"app": "BBB", "phase": "PDT"}}}
	d111 := pb.Data{Signature: &pb.DataSignature{Filename: "testfilename", Md5: "5", Namespace: "toto", PodName: "p111", Metadata: map[string]string{"app": "CCC", "phase": "PDT"}}}
	d2 := pb.Data{Signature: &pb.DataSignature{Filename: "testfilename", Md5: "6", Namespace: "toto", PodName: "p2", Metadata: map[string]string{"app": "AAA", "phase": "UAT"}}}
	d3 := pb.Data{Signature: &pb.DataSignature{Filename: "othertestfilename", Md5: "6", Namespace: "toto", PodName: "p1", Metadata: map[string]string{"app": "AAA", "phase": "UAT"}}}

	inMemStoreInstance.Reference(d1.Signature)
	time.Sleep(time.Second)
	inMemStoreInstance.Reference(d11.Signature)
	time.Sleep(time.Second)
	inMemStoreInstance.Reference(d111.Signature)
	inMemStoreInstance.Reference(d2.Signature)
	inMemStoreInstance.Reference(d3.Signature)

	keys := inMemStoreInstance.GetSameReferences(d1.Signature)
	if len(keys) != 3 {
		t.Errorf("Bad same refcount %#v\n", keys)
	}
	podref := inMemStoreInstance.GetPodReferences("toto", "p1")
	if len(podref) != 2 {
		t.Errorf("Bad podref count %#v\n", keys)
	}
	at := inMemStoreInstance.GetReferencesAt(time.Unix(d1.Signature.GetTimestamp().Seconds, 0), time.Unix(d111.Signature.GetTimestamp().Seconds, 0))
	if len(at) != 2 {
		t.Errorf("Bad at count %#v\n", keys)
	}
	meta := inMemStoreInstance.GetReferencesForMeta("phase", "UAT")
	if len(meta) != 2 {
		t.Errorf("Bad meta count %#v\n", keys)
	}
}
