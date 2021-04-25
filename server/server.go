package main

import (
	"context"
	"log"
	"net"
	//"os"
	"time"

	"google.golang.org/grpc"
	pb "grpcdemo/proto"

	zipkin "github.com/openzipkin/zipkin-go"
	zipkingrpc "github.com/openzipkin/zipkin-go/middleware/grpc"
	zipkinHTTP "github.com/openzipkin/zipkin-go/reporter/http"
	//zipkinlog "github.com/openzipkin/zipkin-go/reporter/log"
)

const (
	port = "localhost:50051"
)

type server struct {
	pb.UnimplementedGreeterServer
}

func (s *server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", req.GetName())
	//span := zipkin.SpanFromContext(ctx)
	//log.Printf("%v", span)

	// db
	newSpan, _ := tracer.StartSpanFromContext(ctx, "mysql_query")
	defer newSpan.Finish()
	//log.Printf("%v", newSpan)
	//log.Printf("%v", newCtx)
	time.Sleep(10 * time.Millisecond)

	// TODO: another grpc call

	return &pb.HelloReply{Message: "Hello " + req.GetName()}, nil
}

var tracer *zipkin.Tracer

func main() {
	log.Printf("start grpc server")
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	//reporter := zipkinlog.NewReporter(log.New(os.Stderr, "", log.LstdFlags))
	reporter := zipkinHTTP.NewReporter("http://localhost:9411/api/v2/spans")
	localEndpoint, err2 := zipkin.NewEndpoint("hello-svr", port)
	if err2 != nil {
		log.Printf("local ep err: %v", err2)
	}
	tracer, _ = zipkin.NewTracer(
		reporter,
		zipkin.WithNoopSpan(true),
		zipkin.WithLocalEndpoint(localEndpoint),
	)

	//s := grpc.NewServer()
	s := grpc.NewServer(grpc.StatsHandler(zipkingrpc.NewServerHandler(tracer)))
	pb.RegisterGreeterServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
