#!/bin/sh
protoc -I kharvest/ kharvest/kharvest.proto --go_out=plugins=grpc:kharvest