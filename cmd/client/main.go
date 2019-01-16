package main

import (
	"context"
	"flag"
	"io"
	"log"

	pb "hilmar/golang_test/pb"

	"google.golang.org/grpc"
	"hilmar/golang_test/internal/pkg/crpt"
	"encoding/base64"
)

var (
	serverAddr  = flag.String("addr", ":8443", "The server address in the format of host:port")
	privetKey   = flag.String("priv-key", "", "Private key")
)

func init() {
	flag.Parse()
}

func main() {
	ctx := context.Background()
	numbers := []int64{1, 5, 3, 6, 2, 20}

	if *privetKey == "" {
		log.Fatalf("you shoud set path to private key via -priv-key param")
	}

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewServerClient(conn)
	stream, err := client.FindMaxNumber(ctx)
	if err != nil {
		log.Fatalf("error while creating stream to server: %v", err)
	}

	waitCh := make(chan struct{})

	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				close(waitCh)
				return
			}
			if err != nil {
				log.Fatalf("failed to receive a note : %v", err)
			}
			log.Printf("Current max: %d", resp.Number)
		}
	}()

	for _, v := range numbers {
		signer, err := crpt.LoadPrivateKey(*privetKey)
		if err != nil {
			log.Printf("signer is damaged: %v", err)
			continue
		}
		signed, err := signer.Sign([]byte(crpt.SIGNATURE_TEXT))
		if err != nil {
			log.Printf("could not sign request: %v", err)
			continue
		}
		sig := base64.StdEncoding.EncodeToString(signed)

		stream.Send(&pb.Request{
			Number: v,
			Sign: sig,
		})
	}

	stream.CloseSend()
	<-waitCh
}
