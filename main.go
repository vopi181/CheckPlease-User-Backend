package main

import (
	"fmt"
	"github.com/vopi181/CheckPlease-User-Backend/CPUser"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
)


func main() {
	log.Println("CheckPlease User Backend Launching at localhost:9000");
	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := CPUser.Server{}

	grpcServer := grpc.NewServer();

	CPUser.RegisterCPUserServer(grpcServer, &s)


	reflection.Register(grpcServer);

	log.Println("Connecting to DB");
	err = CPUser.DBCreateDBConn()
	if err != nil {
		fmt.Errorf("%v", err.Error())
		os.Exit(0);
	}

	err = CPUser.DBPing()
	if err != nil {
		log.Println("Couldnt connect to DB. Make sure a DB is running and creds are correct")
		fmt.Errorf("%v", err.Error())
		os.Exit(0);
	}
	log.Println("Connected to DB")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}



}
