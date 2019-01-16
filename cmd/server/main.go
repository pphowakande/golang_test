package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"time"
	"encoding/base64"
	pb "hilmar/golang_test/pb"
	"google.golang.org/grpc"
	"hilmar/golang_test/internal/pkg/crpt"
)

var (
	port      = flag.Int("port", 8443, "The server port")
	publicKey = flag.String("pub-key", "", "Public key")
)

func init() {
	flag.Parse()
}

func main() {
	if *publicKey == "" {
		log.Fatalf("you shoud set path to public key via -pub-key param")
	}
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterServerServer(grpcServer, &Server{})
	grpcServer.Serve(lis)
}

type Server struct{}

func (s *Server) FindMaxNumber(stream pb.Server_FindMaxNumberServer) error {
	var max *int64
	numCh := make(chan int64, 10)
	waitCh := make(chan struct{})
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				close(numCh)
				return
			}
			if err != nil {
				close(numCh)
				return
			}

			parser, perr := crpt.LoadPublicKey(*publicKey)
			if perr != nil {
				log.Printf("could not sign request: %v", err)
				continue
			}

			sig, err := base64.StdEncoding.DecodeString(in.Sign)
			if err != nil {
				log.Printf("sig base64 decode error")
				continue
			}

			err = parser.Unsign([]byte(crpt.SIGNATURE_TEXT), sig)
			if err != nil {
				log.Printf("could not sign request: %v", err)
				continue
			}

			numCh <- in.Number
		}
	}()

	go func() {
		for {
			select {
			case num, ok := <-numCh:
				if !ok {
					close(waitCh)
					return
				}
				if max == nil || num > *max {
					max = &num
					stream.Send(&pb.Response{
						Number: *max,
					})
				}
			case <-time.After(5 * time.Second):
				close(waitCh)
			}
		}
	}()

	<-waitCh
	return nil
}
