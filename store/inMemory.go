package store

import (
	"sync"

	pb "github.com/dbenque/kharvest/kharvest"
)

type dataPerPod map[string]map[string][]*pb.Data

func newDataPerPod() dataPerPod {
	return map[string]map[string][]*pb.Data{}
}

//InMemory in memroy (map based storage) implementation of store interface
type InMemory struct {
	sync.Mutex
	//files: key1=filename key2=md5 then full data
	files map[string]map[string]pb.Data
	//pods: key1=namespace key2=podName key3=filename then reference to full data
	pods map[string]dataPerPod
}

var _ Store = &InMemory{}

//NewInMemoryStorage create a new InMemoryStorage
func NewInMemoryStorage() *InMemory {
	return &InMemory{
		files: map[string]map[string]pb.Data{},
		pods:  map[string]dataPerPod{},
	}
}

//Has implements Store interface Has method
func (i *InMemory) Has(signature *pb.DataSignature) (bool, bool) {
	mMd5, ok := i.files[signature.GetFilename()]
	if !ok {
		return false, false
	}

	_, dataok := mMd5[signature.GetMd5()]
	return true, dataok
}

//Store implements Store interface Store method
func (i *InMemory) getNamespacePodStore(ns string) dataPerPod {
	if m, ok := i.pods[ns]; ok {
		return m
	}

	m := newDataPerPod()
	i.pods[ns] = m
	return m
}

//return true if the data was added. False it was already there
func (dpp dataPerPod) addData(data *pb.Data) bool {
	found := false
	perFilename, ok := dpp[data.GetSignature().GetPodName()]
	if !ok {
		//create an entry for that podname
		perFilename = map[string][]*pb.Data{}
		dataSlice := []*pb.Data{data}
		perFilename[data.GetSignature().GetFilename()] = dataSlice
	} else {
		if dataSlice, ok := perFilename[data.GetSignature().GetFilename()]; ok {
			//check if that version was already known
			md5 := data.GetSignature().GetMd5()
			for _, v := range dataSlice {
				if v.GetSignature().GetMd5() == md5 {
					found = true
					break
				}
			}
			if !found {
				dataSlice = append(dataSlice, data)
				perFilename[data.GetSignature().GetFilename()] = dataSlice
			}
		}
	}
	dpp[data.GetSignature().GetPodName()] = perFilename
	return !found
}

//Store implements Store interface Store method
func (i *InMemory) Store(data *pb.Data) Action {
	i.Lock()
	defer i.Unlock()

	kf, kv := i.Has(data.GetSignature())
	if kf && kv {
		return None
	}

	mMd5, ok := i.files[data.GetSignature().GetFilename()]
	if !ok { // Completely new storage
		md5Map := map[string]pb.Data{}
		md5Map[data.GetSignature().GetMd5()] = *data
		i.files[data.GetSignature().GetFilename()] = md5Map
		dpp := i.getNamespacePodStore(data.GetSignature().GetNamespace())
		added := dpp.addData(data)
		if added {
			return Create
		}
		return Error
	}

	_, dataok := mMd5[data.GetSignature().GetMd5()]
	if !dataok {
		// This a new version
		mMd5[data.GetSignature().GetMd5()] = *data
		dpp := i.getNamespacePodStore(data.GetSignature().GetNamespace())
		added := dpp.addData(data)
		if added {
			return NewVersion
		}
		return Error
	}

	//Maybe a new reference to that file,version
	dpp := i.getNamespacePodStore(data.GetSignature().GetNamespace())
	added := dpp.addData(data)
	if added {
		return NewVersion
	}
	return None
}
