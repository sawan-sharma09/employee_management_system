// package grpcclient
package grpcServices

// package main

import (
	"context"
	"fmt"
	"io"
	"managedata/grpc_services/pb"
	initpack "managedata/init_pack"
	pubsubops "managedata/pubsub_ops"
	redisops "managedata/redisOps"
	"managedata/util"
)

func Bidi_stream(emp util.Employee) error {
	fmt.Println("Bidirectional streaming service started...")

	var requests []*pb.EmpRequest

	r := &pb.EmpRequest{
		Empdata: &pb.Emp{
			Id:         int32(emp.ID),
			Name:       emp.Name,
			Department: emp.Department,
			Salary:     float32(emp.Salary),
		},
	}

	requests = append(requests, r)

	c := pb.NewEmpServiceClient(initpack.Cc)

	stream, err := c.Bidi_Emp(context.Background())
	if err != nil {
		fmt.Println("Error in client side stream genaration...")
		return err
	}
	waitc := make(chan struct{})

	go func() {
		for _, req := range requests {
			fmt.Println("Sending request: ", req)
			stream.Send(req)
		}
	}()

	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				fmt.Println("All data received")
				// break
				close(waitc)
			}
			if err != nil {
				fmt.Println("Error in receiving stream: ", err)
			}
			fmt.Println("Received-->", res)

			var updatedEmp util.Employee
			updatedEmp.Department = res.Result.Department
			updatedEmp.ID = int(res.Result.Id)
			updatedEmp.Name = res.Result.Name
			updatedEmp.Salary = float64(res.Result.Salary)

			//Update the data in Redis Cache
			redisops.RedisSet(updatedEmp)

			//Publish the updated data to Pubsub topic
			_, publishErr := pubsubops.PublishData(updatedEmp)
			if publishErr != nil {
				fmt.Println("Error in Pubsub publish: ", publishErr)
			}
		}
	}()

	<-waitc
	return nil
}
