package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"

	pb "github.com/dbenque/kharvest/kharvest"
	"github.com/dbenque/kharvest/util"
	google_protobuf "github.com/golang/protobuf/ptypes/empty"
)

const (
	serviceName = "kharvest"
	servicePort = "80"
)

func main() {
	kharvestURL := flag.String("k", "kharvest:80", "URL to kharvest server")
	cmd := flag.String("cmd", "", "Command to be executed")
	podname := flag.String("p", "", "Podname (for cmd=pods)")
	namespace := flag.String("n", "default", "namespace (for cmd=pods)")
	index := flag.Int("i", -1, "index value (for cmd=same)")

	for _, v := range os.Args {
		if v == "-h" || v == "-help" || v == "--h" || v == "--help" || v == "help" {
			printHelp()
			return
		}
	}

	flag.Parse()
	conn, err := grpc.Dial(*kharvestURL, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("[kharvest] [error] grpc client error: %v", err) //Error
	}
	defer conn.Close()
	kharvestUserAPI := pb.NewKharvestUserAPIClient(conn)

	switch *cmd {
	default:
		printHelp()
	case "keys":
		{
			reply, err := kharvestUserAPI.Keys(context.Background(), &google_protobuf.Empty{})
			if err != nil {
				os.Stderr.Write([]byte(err.Error()))
				os.Exit(1)
			}
			for _, v := range reply.Keys {
				fmt.Printf("%s\n", v)
			}
		}
	case "pod", "same":
		{
			if *podname == "" {
				os.Stderr.Write([]byte("Missing podname, use parameter 'p'"))
				os.Exit(1)
			}

			reply, err := kharvestUserAPI.PodReferences(context.Background(), &pb.PodIdentifier{Namespace: *namespace, PodName: *podname})
			if err != nil {
				os.Stderr.Write([]byte(err.Error()))
				os.Exit(1)
			}

			if *cmd == "pod" {
				for i, s := range reply.Signatures {
					fmt.Printf("[%d] %s with sign %s at %s\n", i, s.Filename, util.MD5toStr64(s), time.Unix(s.GetTimestamp().Seconds, 0).Format(time.UnixDate))
				}
				return
			}

			if *cmd == "same" {
				if *index == -1 {
					os.Stderr.Write([]byte("Missing index, use parameter 'i'"))
					os.Exit(1)
				}
				if *index < 0 || *index >= len(reply.Signatures) {
					os.Stderr.Write([]byte("Index out of bound. Use command 'pods'."))
					os.Exit(1)
				}

				sameSignature := reply.Signatures[*index]
				reply, err := kharvestUserAPI.SameReferences(context.Background(), sameSignature)
				if err != nil {
					os.Stderr.Write([]byte(err.Error()))
					os.Exit(1)
				}

				fmt.Printf("File %s with sign %s is referenced by the following pods:\n", sameSignature.Filename, util.MD5toStr64(sameSignature))

				for _, s := range reply.Signatures {
					fmt.Printf("    %s/%s at %s\n", s.Namespace, s.PodName, time.Unix(s.GetTimestamp().Seconds, 0).Format(time.UnixDate))
				}
			}
		}

	}
}

func printHelp() {
	fmt.Printf("k         : url for the kharvest server [kharvest:80]\n")
	fmt.Printf("help      : display this help\n")
	fmt.Printf("cmd       : command to be excuted in the following list:\n")
	fmt.Printf(" --> pod      : list all the file referenced by a pod\n")
	fmt.Printf("     parameters: p=[podname] n=[namespace]\n")
	fmt.Printf(" --> same      : list all the pods that reference the same file reference. The index is the one returned by the 'pods' command.\n")
	fmt.Printf("     parameters: p=[podname] n=[namespace] i=[index]\n")
	fmt.Println("")
	fmt.Println("Example: ")
	fmt.Println(os.Args[0], " -k=192.168.99.100:32140 -cmd=pod -p=kharvestclient-244272239-2ncxr")
	fmt.Println(os.Args[0], " -k=192.168.99.100:32140 -cmd=same -p=kharvestclient-244272239-2ncxr -i=1")

}
