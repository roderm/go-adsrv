package grpc_server

import (
	"context"
	"fmt"
	pb "github.com/roderm/go-adsrv/api/proto/go/address"
	as "github.com/roderm/go-adsrv/api/proto/go/address-service"
	"github.com/roderm/pubsub"
	"strings"
)

type address struct {
	get pubsub.Topic
	set pubsub.Topic
	v   *pb.Value
}
type addr struct {
	reg *Registry
}

func NewAddrSrv(r *Registry) as.AddressServer {
	return &addr{
		reg: r,
	}
}

func (a *addr) Get(_ *pb.Empty, stream as.Address_GetServer) error {
	return nil
}
func (a *addr) Set(ctx context.Context, val *pb.Value) (*pb.Value, error) {
	addr := strings.Split(val.GetAddress(), ":")
	err := a.reg.Set(
		addr[0],
		strings.Join(addr[1:], ":"),
		val.GetValue(),
	)
	return val, err
	// TODO: sub-once for...
}

func (a *addr) Subscribe(sub *pb.AddressSub, pub as.Address_SubscribeServer) error {
	ch, err := a.reg.ps.Sub(pub.Context(), sub.Ids...)
	if err != nil {
		return err
	}
	for {
		select {
		case <-pub.Context().Done():
			return pub.Context().Err()
		case msg := <-ch:
			addr, ok := msg.(*pb.AddressObject)
			if !ok {
				fmt.Println("invalid type published")
				continue
			}
			err := pub.Send(&pb.Value{
				Address: addr.Address,
				Value:   addr.Value,
			})

			if err != nil {
				return err
			}
		}
	}
}
