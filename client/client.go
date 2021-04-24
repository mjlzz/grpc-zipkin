package main

import (
	"context"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"

	zipkin "github.com/openzipkin/zipkin-go"
	zipkingrpc "github.com/openzipkin/zipkin-go/middleware/grpc"
	zipkinmodel "github.com/openzipkin/zipkin-go/model"
	zipkinHTTP "github.com/openzipkin/zipkin-go/reporter/http"
	//zipkinlog "github.com/openzipkin/zipkin-go/reporter/log"
)

const (
	address     = "localhost:50051"
	defaultName = "world"
)

func main() {
	// reporter := zipkinlog.NewReporter(log.New(os.Stderr, "", log.LstdFlags))
	reporter := zipkinHTTP.NewReporter("http://localhost:9411/api/v2/spans")
	tracer, _ := zipkin.NewTracer(
		reporter,
		zipkin.WithNoopSpan(true),
	)

	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithStatsHandler(zipkingrpc.NewClientHandler(tracer)))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	endpoint, err2 := zipkin.NewEndpoint("helloService", address)
	if err2 != nil {
		log.Printf("new endpoint err", err2)
	}
	//span := tracer.StartSpan("client_request2")
	span, newCtx := tracer.StartSpanFromContext(ctx, "client_request", zipkin.RemoteEndpoint(endpoint), zipkin.Kind(zipkinmodel.Client))
	//defer span.Finish() // defer will exit quickly, reporter hasn't finish sending
	log.Printf("%v", newCtx)
	
	r, err := c.SayHello(newCtx, &pb.HelloRequest{Name: defaultName})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	span.Finish()

	log.Printf("Greeting: %s", r.GetMessage())
	
	time.Sleep(time.Second) // wait for a while, let reporter finish sending
}
