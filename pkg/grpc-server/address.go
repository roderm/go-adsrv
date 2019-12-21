package grpc_server

import (
	"context"
	"strings"

	pb "github.com/roderm/go-adsrv/api/proto/go/address"
	as "github.com/roderm/go-adsrv/api/proto/go/address-service"
	log "github.com/sirupsen/logrus"
)

type address struct {
	v *pb.Value
}
type addr struct {
	reg *Registry
}

// NewAddrSrv creating a Server to Set/Get values
func NewAddrSrv(r *Registry) as.AddressServer {
	return &addr{
		reg: r,
	}
}

func (a *addr) Get(sub *pb.AddressSub, stream as.Address_GetServer) error {
	for _, id := range sub.GetIds() {
		addr, err := a.reg.get(id)
		if err != nil {
			return err
		}
		err = stream.Send(addr)
		if err != nil {
			return err
		}
	}
	return nil
}

// Set sends an AddressObject to the Plugin
func (a *addr) Set(ctx context.Context, val *pb.Value) (*pb.Value, error) {
	addr := strings.Split(val.GetAddress(), ":")
	err := a.reg.Set(
		addr[0],
		strings.Join(addr[1:], ":"),
		val.GetValue(),
	)
	return val, err
}

// Subscribe to Addresses from the Plugins
func (a *addr) Subscribe(sub *pb.AddressSub, pub as.Address_SubscribeServer) error {
	ch, err := a.reg.ps.Sub(pub.Context(), sub.Ids...)
	if err != nil {
		log.Warnf("Failed opening sub-stream: %v", err)
		return err
	}
	for {
		select {
		case <-pub.Context().Done():
			log.Warnf("publisher exiting context")
			return pub.Context().Err()
		case msg := <-ch:
			addr, ok := msg.(*pb.AddressObject)
			if !ok {
				log.Warn("invalid type published")
				continue
			}
			err := pub.Send(&pb.Value{
				Address: addr.Address,
				Value:   addr.Value,
			})

			if err != nil {
				log.Warnf("Couldn't read value: %v", err)
				return err
			}
		}
	}
}
