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
	// check if we pass public key in arguments
	if *publicKey == "" {
		log.Fatalf("you shoud set path to public key via -pub-key param")
	}
	// creating server listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption

	// creatibe new server using grpc model
	grpcServer := grpc.NewServer(opts...)
	// registering newly create service using GRPC implementation
	pb.RegisterServerServer(grpcServer, &Server{})
	// serve accepts incoming connections on listener lis
	grpcServer.Serve(lis)
}

type Server struct{}

func (s *Server) FindMaxNumber(stream pb.Server_FindMaxNumberServer) error {
	var max *int64
	// using go channels here to send and receive values
	// go channel => pipes that connect concurrent goroutines
	// we can send values into channels from one goroutine and receive them into another goroutine
	numCh := make(chan int64, 10)
	waitCh := make(chan struct{})
	go func() {
		for {
			// receiving stream values
			in, err := stream.Recv()
			// error handling logic
			if err == io.EOF {
				close(numCh)
				return
			}
			if err != nil {
				close(numCh)
				return
			}

			// loading pubic key
			//LoadPublicKey parses PEM encoded public key file
			parser, perr := crpt.LoadPublicKey(*publicKey)
			if perr != nil {
				log.Printf("could not sign request: %v", err)
				continue
			}

			// decoding signature from client
			sig, err := base64.StdEncoding.DecodeString(in.Sign)
			if err != nil {
				log.Printf("sig base64 decode error")
				continue
			}

			//Unsign verifies the message using a rsa-sha256 signature
			err = parser.Unsign([]byte(crpt.SIGNATURE_TEXT), sig)
			if err != nil {
				log.Printf("could not sign request: %v", err)
				continue
			}

			// writing to channel
			numCh <- in.Number
		}
	}()

	go func() {
		for {
			// Combining goroutines and channels using select
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
