package grpc_server

import (
	"context"
	"errors"
	"fmt"
	pb "github.com/roderm/go-adsrv/api/proto/go/address"
	rs "github.com/roderm/go-adsrv/api/proto/go/registry-service"
	"github.com/roderm/go-adsrv/pkg/pubsub"
)

// Registry keeps the registered Plugins from gRPC
// Implements rs.RegistryServer
// handles receive and send
type Registry struct {
	ctx   context.Context
	store map[interface{}]*grpcPlugable
	ps    *pubsub.Pubsub
}
type grpcPlugable struct {
	conn      pubPlugin
	addresses map[string]*pb.AddressObject
}
type pubPlugin interface {
	Send(*pb.Value) error
}

// NewRegistry receive gRPC-Service
func NewRegistry(ctx context.Context) *Registry {
	r := &Registry{
		ctx:   ctx,
		store: make(map[interface{}]*grpcPlugable),
		ps:    pubsub.NewPubsub(),
	}
	return r
}

// Register handles incoming Plugins
func (r *Registry) Inform(srv rs.Registry_InformServer) error {
	srvId := srv.Context().Value("id")
	if _, ok := r.store[srvId]; ok {
		return errors.New("Plugin already exists")
	}
	plugable := &grpcPlugable{
		conn:      srv,
		addresses: make(map[string]*pb.AddressObject),
	}
	r.store[srvId] = plugable
	defer func() {
		delete(r.store, srv)
	}()
	for {
		select {
		case <-r.ctx.Done():
			return r.ctx.Err()
		case <-srv.Context().Done():
			return r.ctx.Err()
		default:
			v, err := srv.Recv()
			if err != nil {
				return err
			}
			address := v.GetAddress()
			v.Address = fmt.Sprintf("%s:%s", srvId, v.GetAddress())
			plugable.addresses[address] = v
			// store type
			r.ps.Publish("new_address", v.Address)
		}
	}
}

func (r *Registry) ValueUpdate(srv rs.Registry_ValueUpdateServer) error {
	defer srv.SendAndClose(nil)
	srvId := srv.Context().Value("id")
	var plugable *grpcPlugable
	var ok bool
	if plugable, ok = r.store[srvId]; !ok {
		return errors.New("Plugin not loaded")
	}
	for {
		select {
		case <-r.ctx.Done():
			return r.ctx.Err()
		case <-srv.Context().Done():
			return r.ctx.Err()
		default:
			v, err := srv.Recv()
			if err != nil {
				return err
			}
			plugable.addresses[v.GetAddress()].Value = v.Value
			address := fmt.Sprintf("%s:%s", srvId, v.GetAddress())
			r.ps.Publish(plugable.addresses[v.GetAddress()], address)

			fmt.Printf("New value %v on address %s \n", v.GetValue(), address)
		}
	}
}

// Set sends Set-command to plugin
func (r *Registry) Set(srvId interface{}, address string, value []byte) error {
	var srv *grpcPlugable
	var ok bool
	if srv, ok = r.store[srvId]; !ok {
		return errors.New("Has been closed or never existed")
	}
	return srv.conn.Send(&pb.Value{
		Address: address,
		Value:   value,
	})
}
