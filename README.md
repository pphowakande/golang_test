# GRPC bidirectional stream
Example of implementation grpc bidirectional stream with request key verification.

## Short description
Implement a FindMaxNumber RPC Bi-Directional Streaming Client and Server
system:

1. The function takes a stream of a Request message that has one integer, and
returns a stream of Responses that represent the current maximum between
all these integers

2. Client will be having a cryptographic public key and client will be identified
using his private key. The client will sign every request message in the stream.

3. Each requested message should be verified against the signature at the
server end. Only those numbers will be considered to be processed whose
sign is successfully verified.

Example: The client will send a stream of number (1,5,3,6,2,20) and each number
will be signed by the private key of the client and the server will respond with a
stream of numbers (1,5,6,20).

## Project structure
/cert - example of public
/cmd/server - server
/cmd/client - client
/internal/pkg/crpt - package for sign/unsign message with public/private key
/pb - service proto file and generate based on proto file code

## Dependencies
    github.com/golang/protobuf v1.2.0
    google.golang.org/grpc v1.18.0
    golang.org/x/net

## Server
*run help:*
```
go run cmd/server/main.go -h
```
*example of running server:*
```
go run cmd/server/main.go -pub-key cert/public.pem
```
## How to run client
*run help:*
```
go run cmd/client/main.go -h
```
*example of running client:*
```
go run cmd/client/main.go  -priv-key cert/public.pem
```

## Custom key generation
### Private
```
openssl genrsa -out private.pem 2048
```
### Public
```
openssl rsa -in private.pem -outform PEM -pubout -out public.pem
```
