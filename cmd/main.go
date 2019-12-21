package main

import (
	"context"
	"fmt"
	"github.com/micro/cli"
	as "github.com/roderm/go-adsrv/api/proto/go/address-service"
	rs "github.com/roderm/go-adsrv/api/proto/go/registry-service"
	"github.com/roderm/go-adsrv/pkg/grpc-server"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"os"
)

func main() {

	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:   "port",
			Usage:  "Port on which the gateway runs",
			Value:  3000,
			EnvVar: "PORT",
		},
	}

	app.Action = func(c *cli.Context) error {
		n := fmt.Sprintf("%s:%d", "localhost", c.Int("port"))

		srv := grpc.NewServer()
		reg := grpc_server.NewRegistry(context.Background())
		addvrk := grpc_server.NewAddrSrv(reg)

		rs.RegisterRegistryServer(srv, reg)
		as.RegisterAddressServer(srv, addvrk)
		l, err := net.Listen("tcp", n)
		if err != nil {
			return err
		}

		log.Infof("Running on %s", n)
		defer l.Close()
		return srv.Serve(l)
	}
	panic(app.Run(os.Args))
}
