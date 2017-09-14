package util

import (
	"encoding/base64"
	"fmt"

	pb "github.com/dbenque/kharvest/kharvest"
)

//BuildKeyString generate a unique string from the DataSignature
func BuildKeyString(dataSignature *pb.DataSignature) string {
	if dataSignature == nil {
		return ""
	}
	return dataSignature.Filename + "." + dataSignature.Md5
}

func DefaultValueForSignature(dataSignature *pb.DataSignature) error {

	if dataSignature == nil {
		return fmt.Errorf("nil value for datasignature")
	}

	if len(dataSignature.Filename) == 0 {
		return fmt.Errorf("Missing filename for datasignature: %v", dataSignature)
	}

	if len(dataSignature.PodName) == 0 {
		return fmt.Errorf("Missing podname for datasignature: %v", dataSignature)
	}

	if len(dataSignature.Namespace) == 0 {
		return fmt.Errorf("Missing namespace for datasignature: %v", dataSignature)
	}

	if len(dataSignature.Md5) == 0 {
		return fmt.Errorf("Missing MD5 for datasignature: %v", dataSignature)
	}

	if dataSignature.Metadata == nil {
		dataSignature.Metadata = map[string]string{"None": "None"}
	}
	return nil
}

func MD5toStr64(dataSignature *pb.DataSignature) string {
	return base64.StdEncoding.EncodeToString([]byte(dataSignature.GetMd5()))
}
