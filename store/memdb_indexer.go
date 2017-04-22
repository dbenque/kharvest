package store

import (
	"fmt"
	"strconv"

	pb "github.com/dbenque/kharvest/kharvest"
	"github.com/golang/protobuf/ptypes/timestamp"
)

type hasTimestamp interface {
	GetTimestamp() *timestamp.Timestamp
}
type timestampIndexer struct {
	modSeconds int64
}

func (s *timestampIndexer) FromObject(obj interface{}) (bool, []byte, error) {
	data, ok := obj.(hasTimestamp)
	if !ok {
		return false, nil, fmt.Errorf("Not of type hasTimestamp")
	}

	t := data.GetTimestamp()
	if t == nil {
		return false, nil, fmt.Errorf("Timestamp not set")
	}
	tSec := t.Seconds
	if s.modSeconds > 1 {
		tSec = tSec - tSec%s.modSeconds
	}
	val := strconv.FormatInt(tSec, 36)
	val += "\x00"
	return true, []byte(val), nil
}

func (s *timestampIndexer) FromArgs(args ...interface{}) ([]byte, error) {
	return fromArgs(args...)
}

type idDataIndexer struct {
	keyBuilderFunction KeyBuilderFunction
}

func (s *idDataIndexer) FromObject(obj interface{}) (bool, []byte, error) {
	dataSignature, ok := obj.(*pb.DataSignature)
	if !ok {
		data, ok := obj.(*pb.Data)
		if !ok {
			return false, nil, fmt.Errorf("Unknown object type. Not pb.Data nor pb.DataSignature")
		}
		return s.FromObject(data.Signature)
	}

	val := s.keyBuilderFunction(dataSignature)
	val += "\x00"
	return true, []byte(val), nil
}

func (s *idDataIndexer) FromArgs(args ...interface{}) ([]byte, error) {
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
