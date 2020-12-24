package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	pb "tt47cf6/minecraft/protos"

	"google.golang.org/grpc"
)

type RPCServer struct {
	pb.UnimplementedMyRPCServerServer

	s   *grpc.Server
	smu sync.Mutex
}

func NewRPCServer() *RPCServer {
	return &RPCServer{}
}

func (s *RPCServer) BlockingServe(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("net.Listen(%d): %v", port, err)
	}

	s.smu.Lock()
	s.s = grpc.NewServer()
	s.smu.Unlock()

	pb.RegisterMyRPCServerServer(s.s, s)

	log.Printf("starting rpc server on port %d", port)
	return s.s.Serve(lis)
}

func (s *RPCServer) Stop(ctx context.Context) error {
	s.smu.Lock()
	defer s.smu.Unlock()
	if s.s == nil {
		return nil
	}

	doneChan := make(chan bool)
	go func() {
		s.s.GracefulStop()
		doneChan <- true
	}()

	select {
	case <-doneChan:
		return nil
	case <-ctx.Done():
		log.Print("force-stopping the RPC server")
		s.s.Stop()
	}

	return nil
}

func (s *RPCServer) Dummy(ctx context.Context, cfg *pb.DummyRequest) (*pb.SimpleResponse, error) {
	return &pb.SimpleResponse{
		Success: true,
		Message: "Hello world!",
	}, nil
}
