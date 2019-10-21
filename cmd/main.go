package main

import (
	"context"
	"fmt"
	as "github.com/roderm/go-adsrv/api/proto/go/address-service"
	rs "github.com/roderm/go-adsrv/api/proto/go/registry-service"
	"github.com/roderm/go-adsrv/pkg/grpc-server"
	"google.golang.org/grpc"
	"net"
)

func main() {
	l, err := net.Listen("unix", "/tmp/adsrv.sock")
	if err != nil {
		panic(err)
	}
	fmt.Println("listeing on unix")
	// get services
	reg := grpc_server.NewRegistry(context.Background())
	fmt.Println("services")
	addvrk := grpc_server.NewAddrSrv(reg)
	srv := grpc.NewServer()
	fmt.Println("register")
	rs.RegisterRegistryServer(srv, reg)
	as.RegisterAddressServer(srv, addvrk)
	fmt.Println("start server")
	defer l.Close()
	panic(srv.Serve(l))
}
