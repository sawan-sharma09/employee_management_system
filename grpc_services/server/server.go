package main

import (
	"fmt"
	"grpc_new_server/conn"
	"grpc_new_server/db"
	"grpc_new_server/grpc_services/pb"
	"grpc_new_server/grpc_services/utils"
	"io"

	"google.golang.org/grpc"
)

type server struct{}

func main() {
	fmt.Println("--gRPC server--")
	fmt.Println("Bidirectional streaming services started...")

	s := grpc.NewServer()
	pb.RegisterEmpServiceServer(s, &server{})

	if err := s.Serve(conn.Lis); err != nil {
		fmt.Println("Error in listening: ", err)
	}
}

func (*server) Bidi_Emp(stream pb.EmpService_Bidi_EmpServer) error {

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("All data received..")
			break
		}
		if err != nil {
			fmt.Println("Error while receiving data from stream..", err)
			return err
		}

		department := req.GetEmpdata().GetDepartment()
		id := req.GetEmpdata().GetId()
		name := req.GetEmpdata().GetName()
		salary := req.GetEmpdata().GetSalary()

		empData := &utils.Employee{Name: name, Department: department, ID: int(id), Salary: float64(salary)}
		fmt.Println("Data from client: ", *empData)

		//check the Employee's existance and update if exists
		updatedEmp, dbErr := db.DbUpdateEmp(empData)
		if dbErr != nil {
			fmt.Println(dbErr)
			// return dbErr
		}

		//Store updated data in empDetails for response
		empDetails := &pb.Emp{Name: updatedEmp.Name, Department: updatedEmp.Department, Id: int32(updatedEmp.ID), Salary: float32(updatedEmp.Salary)}
		fmt.Println("Updated details: ", empDetails)

		streamErr := stream.Send(&pb.EmpResponse{Result: empDetails})
		if streamErr != nil {
			fmt.Println("Error in server stream...", streamErr)
			return streamErr
		}
	}

	return nil
}

func (*server) MustEmbedUnimplementedEmpServiceServer() {
	panic("Unimplemented...")
}
