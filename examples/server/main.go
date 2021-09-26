package main

import (
	"context"
	"github.com/SkyAPM/go2sky/reporter"
	"google.golang.org/grpc"
	"log"
	"net"
	"sky2gorpc/examples"
	"sky2gorpc/grpc/interceptors"
)

type server struct{}

func (s *server) Greet(ctx context.Context, in *examples.HelloReq) (*examples.HelloResp, error) {
	return &examples.HelloResp{Greet: "Hello " + in.Name}, nil
}

func main() {
	re, err := reporter.NewGRPCReporter("localhost:11800")
	if err != nil {
		log.Fatalf("new reporter error %v \n", err)
	}
	defer re.Close()

	h, err := interceptors.NewTracerHandler(re, "bookstore-api")

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer(grpc.UnaryInterceptor(h.RPCServerTracingInterceptor(h.GetTracker())))
	examples.RegisterHelloGreeterServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
