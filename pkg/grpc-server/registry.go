package grpc_server

import (
	"context"
	"fmt"

	pb "github.com/roderm/go-adsrv/api/proto/go/address"
	rs "github.com/roderm/go-adsrv/api/proto/go/registry-service"
	"github.com/roderm/go-adsrv/pkg/pubsub"
	log "github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
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

func (r *Registry) get(address string) (*pb.AddressObject, error) {
	for _, plug := range r.store {
		addr, ok := plug.addresses[address]
		if ok {
			return addr, nil
		}
	}
	return nil, xerrors.Errorf("Address %s not found in any plugin", address)
}

//Inform handles incoming Plugins
func (r *Registry) Inform(srv rs.Registry_InformServer) error {
	srvID := srv.Context().Value("id")
	log.Debug("New plugin registered.")
	if _, ok := r.store[srvID]; ok {
		return xerrors.New("Plugin already exists")
	}
	plugable := &grpcPlugable{
		conn:      srv,
		addresses: make(map[string]*pb.AddressObject),
	}
	r.store[srvID] = plugable
	defer func() {
		delete(r.store, srv)
	}()
	for {
		select {
		case <-r.ctx.Done():
			return r.ctx.Err()
		case <-srv.Context().Done():
			log.Info("Plugin is gone....")
			return r.ctx.Err()
		default:
			v, err := srv.Recv()
			if err != nil {
				return err
			}
			address := v.GetAddress()
			v.Address = fmt.Sprintf("%s:%s", srvID, v.GetAddress())
			plugable.addresses[address] = v
			// store type
			r.ps.Publish("new_address", v.Address)
		}
	}
}

//ValueUpdate handling updates for the plugin
func (r *Registry) ValueUpdate(srv rs.Registry_ValueUpdateServer) error {
	defer srv.SendAndClose(nil)
	srvID := srv.Context().Value("id")
	var plugable *grpcPlugable
	var ok bool
	if plugable, ok = r.store[srvID]; !ok {
		return xerrors.New("Plugin not loaded")
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
			address := fmt.Sprintf("%s:%s", srvID, v.GetAddress())
			r.ps.Publish(plugable.addresses[v.GetAddress()], address)

			log.Debugf("New value %v on address %s \n", v.GetValue(), address)
		}
	}
}

// Set sends Set-command to plugin
func (r *Registry) Set(srvID interface{}, address string, value []byte) error {
	var srv *grpcPlugable
	var ok bool
	if srv, ok = r.store[srvID]; !ok {
		return xerrors.New("Has been closed or never existed")
	}
	return srv.conn.Send(&pb.Value{
		Address: address,
		Value:   value,
	})
}
