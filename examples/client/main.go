package main

import (
	"context"
	"github.com/SkyAPM/go2sky/reporter"
	"google.golang.org/grpc"
	"log"
	"sky2gorpc/examples"
	"sky2gorpc/grpc/interceptors"
	"time"
)

const (
	address = "localhost:50051"
)

func main() {
	re, err := reporter.NewGRPCReporter("localhost:11800")
	if err != nil {
		log.Fatalf("new reporter error %v \n", err)
	}
	defer re.Close()

	h, err := interceptors.NewTracerHandler(re, "hello-api")
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithUnaryInterceptor(h.RPCClientTracingInterceptor))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer func() {
		time.Sleep(time.Second * 10)
		conn.Close()
	}()
	c := examples.NewHelloGreeterClient(conn)

	// Contact the server and print out its response.
	name := "Apple iPhone 11"

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.Greet(ctx, &examples.HelloReq{Name: name})
	if err != nil {
		log.Fatalf("Could not greeting: %v", err)
	}

	log.Printf("%s", r)
}
