package main

import (
	"context"
	fmt "fmt"
	"net"
	pb "route_chat/route_chat"
	"sync"

	"google.golang.org/grpc"
)

// ConnectionClient is used to creat struct saving client connection
type ConnectionClient struct {
	stream pb.Broadcast_CreateStreamServer
	id     string
	displayName string
	active bool
	err    chan error
}

// Server is the struct to create server
type Server struct {
	Connection []*ConnectionClient
}

// CreateStream is the function to do
func (s *Server) CreateStream(pconn *pb.Connect, stream pb.Broadcast_CreateStreamServer) error {
	conn := &ConnectionClient{
		stream: stream,
		id:     pconn.User.Id,
		displayName: pconn.User.DisplayName,
		active: true,
		err:    make(chan error),
	}
	s.Connection = append(s.Connection, conn)
	return <-conn.err
}

// BroadcastMessage is used to broadcast message to other client
func (s *Server) BroadcastMessage(ctx context.Context, mess *pb.Message) (*pb.Close, error) {
	wait := sync.WaitGroup{}
	done := make(chan int)

	for _, conn := range s.Connection {

		wait.Add(1)
		go func(mess *pb.Message, conn *ConnectionClient) {
			messType:=mess.MessType
			if conn.active {

				if messType =="private"{
				receiverName:=mess.ReceiverDisplayName
				if receiverName==conn.displayName{
					err:=conn.stream.Send(mess)
					fmt.Printf("Sending message %v to user %v ", mess.Id, conn.id)
					if err != nil {
						fmt.Printf("error with stream %v ", conn.stream)
						conn.active = false
					}

				}

				} else if messType =="public"{	
					err := conn.stream.Send(mess)
					fmt.Printf("Sending message %v to user %v ", mess.Id, conn.id)
					if err != nil {
						fmt.Printf("error with stream %v ", conn.stream)
						conn.active = false
					}
					}
			}
			defer wait.Done()
			
		}(mess, conn)

	}
	go func() {
		wait.Wait()
		close(done)
	}()
	<-done
	return &pb.Close{}, nil
}
func main() {
	port := ":8080"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Print("Failed to listen: ", err)
	} else {
		fmt.Printf("listening on localhost:%v", port)
	}

	s := grpc.NewServer()

	pb.RegisterBroadcastServer(s, &Server{})
	if err := s.Serve(lis); err != nil {
		fmt.Print("Fail to serve:", err)
	} else {
		fmt.Printf("listening on localhost:%v", port)
	}

}
