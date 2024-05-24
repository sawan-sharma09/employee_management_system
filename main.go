package main

import (
	"fmt"
	"grpc_new_server/conn"
	"grpc_new_server/grpc_services/pb"
	grpcserver "grpc_new_server/grpc_services/server"

	"google.golang.org/grpc"
)

func main() {
	fmt.Println("--gRPC server--")
	fmt.Println("Bidirectional streaming services started...")

	s := grpc.NewServer()
	var server grpcserver.Server
	pb.RegisterEmpServiceServer(s, &server)

	if err := s.Serve(conn.Lis); err != nil {
		fmt.Println("Error in listening: ", err)
	}
}
