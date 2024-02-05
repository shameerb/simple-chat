package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	google_protobuf "github.com/golang/protobuf/ptypes/empty"
	pb "github.com/shameerb/tcp-chat-redis/pkg/grpcapi"
	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedChatServiceServer
	redisAddr  string
	redis      *redis
	grpcPort   string
	listener   net.Listener
	grpcServer *grpc.Server
	// ctx        context.Context
	// cancel     context.CancelFunc
	// wg         sync.WaitGroup
}

func NewServer(redisAddr, grpcPort string) *Server {
	// ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		redisAddr: redisAddr,
		grpcPort:  grpcPort,
		// ctx:       ctx,
		// cancel:    cancel,
	}
}

func (s *Server) Run() error {
	var err error
	s.redis, err = initRedis(s.redisAddr)
	if err != nil {
		return fmt.Errorf("failed to initialize redis client: %s", err)
	}
	log.Println("initialized redis")
	//todo: Do you need this when you are already closing in the stop function.
	// defer s.redis.client.Close()

	if err = s.startGrpcServer(); err != nil {
		return fmt.Errorf("failed to start grpc server %s", err)
	}
	log.Println("initialized gRPC server")

	// This is called on OS interrupts close anyway
	// defer s.closeGrpcConnection()

	// hold the main server routine until an interrupt occurs.
	s.awaitShutdown()
	return nil
}

func (s *Server) startGrpcServer() error {
	// start grpc server
	var err error
	s.listener, err = net.Listen("tcp", fmt.Sprintf(":%s", s.grpcPort))
	if err != nil {
		return fmt.Errorf("listener failed to initialize: %s", err)
	}
	s.grpcServer = grpc.NewServer()
	// todo: Ideally create a server of chatserviceserver
	pb.RegisterChatServiceServer(s.grpcServer, s)
	// run a goroutine to serve on the grpc server. You can just do a grpcServer.Serve(), but a goroutine helps in initializing other things apart from a grpc server as well and doesnt hold the main routine.
	go func() {
		if err := s.grpcServer.Serve(s.listener); err != nil {
			// todo: this will shutdown the service. Instead you can send a done on the ctx to close all the connections and wait for processing to be done.
			log.Fatalf("Grpc server cannot serve: %s", err)
		}
	}()
	return nil
}

func (s *Server) closeGrpcConnection() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			log.Printf("Error while closing listener: %s", err)
		}
	}
}

func (s *Server) awaitShutdown() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	// wait until you get an interrupt signal
	<-stop
	s.stop()
}

func (s *Server) stop() {
	// call cancel for the context
	log.Println("Stopping server..")
	// s.cancel()
	s.closeGrpcConnection()
	s.redis.client.Close()
}

func (s *Server) Connect(ctx context.Context, req *pb.ConnectRequest) (*google_protobuf.Empty, error) {
	key := "active." + req.GetUser()
	if s.redis.exists(key) {
		return nil, errors.New("user already exists and is connected. choose another username")
	}
	if !s.redis.set(key, req.GetUser()) {
		return nil, errors.New("user already exists and is connected. choose another username")
	}
	if err := s.redis.publish(req.GetUser() + " connected."); err != nil {
		return nil, errors.New("could not publish the user connected message")
	}
	log.Println(req.GetUser() + " connected.")
	return &google_protobuf.Empty{}, nil

}

func (s *Server) Chat(ctx context.Context, msg *pb.Message) (*google_protobuf.Empty, error) {
	key := "active." + msg.GetUser()
	if !s.redis.exists(key) {
		return nil, errors.New("user already exists and is connected. choose another username")
	}
	pubMsg := fmt.Sprintf("> %s : %s", msg.GetUser(), msg.GetMsg())
	if err := s.redis.publish(pubMsg); err != nil {
		return nil, err
	}
	return &google_protobuf.Empty{}, nil
}

func (s *Server) ListUsers(ctx context.Context, in *google_protobuf.Empty) (*pb.UserListResponse, error) {
	res, err := s.redis.search("active.*")
	if err != nil {
		return nil, err
	}
	return &pb.UserListResponse{
		User: res,
	}, nil
}

func (s *Server) Disconnect(ctx context.Context, req *pb.DisconnectRequest) (*google_protobuf.Empty, error) {
	key := "active." + req.GetUser()
	if s.redis.exists(key) {
		if err := s.redis.publish(fmt.Sprintf("> %s, left the chat", req.GetUser())); err != nil {
			return nil, err
		}
		if err := s.redis.delete(key); err != nil {
			return nil, err
		}
	}
	log.Printf("%s disconnected !", req.GetUser())
	return &google_protobuf.Empty{}, nil
}
