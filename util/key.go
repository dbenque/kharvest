package util

import pb "github.com/dbenque/kharvest/kharvest"

//BuildKeyString generate a unique string from the DataSignature
func BuildKeyString(dataSignature *pb.DataSignature) string {
	if dataSignature == nil {
		return ""
	}
	return dataSignature.Filename + "." + dataSignature.Md5
}
