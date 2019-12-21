package client

import (
	"context"
	"errors"
	"github.com/roderm/go-adsrv/api/proto/go/address"
	"github.com/roderm/go-adsrv/api/proto/go/registry-service"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type notifier struct {
	address string
	vals    chan address.Value
}

// Plugin for go-adsrv
type Plugin struct {
	Ctx     context.Context
	cl      registry_service.RegistryClient
	in      registry_service.Registry_InformClient
	vu      registry_service.Registry_ValueUpdateClient
	setters map[string]Settable
	readers map[string]Readable
	vals    chan address.Value
}

// Value send value to the notification
func (n *notifier) Value(v []byte) {
	n.vals <- address.Value{
		Address: n.address,
		Value:   v,
	}
}

// NewPlugin returns a Plugin
func NewPlugin(ctx context.Context, conn *grpc.ClientConn) (*Plugin, error) {
	cl := registry_service.NewRegistryClient(conn)
	log.Debug("Registered to server")
	in, err := cl.Inform(ctx)
	log.Debug("Registering as plugin")
	if err != nil {
		return nil, err
	}
	vu, err := cl.ValueUpdate(ctx)
	if err != nil {
		return nil, err
	}
	p := &Plugin{
		setters: make(map[string]Settable),
		vals:    make(chan address.Value, 100),
		Ctx:     ctx,
		cl:      cl,
		in:      in,
		vu:      vu,
	}

	go p.observeGetter()
	go p.observeSetter()
	return p, nil
}

// AddAddress adding an Address-Object to the Plugin
func (p *Plugin) AddAddress(a IAdress) error {
	ao := a.GetObject()
	if ao.Readable {
		r, ok := a.(Readable)
		// TODO: note about
		if !ok {
			return errors.New("Can't add readable that doesn't implement Readable")
		}
		r.Notifier(&notifier{
			address: ao.Address,
			vals:    p.vals,
		})
	}
	if ao.Writable {
		w, ok := a.(Settable)
		if !ok {
			return errors.New("Can't add writable that doesn't implement Settable")
		}
		p.setters[ao.GetAddress()] = w
	}
	return nil
}
func (p *Plugin) observeSetter() {
	for {
		select {
		case <-p.Ctx.Done():
			return
		default:
			v, _ := p.in.Recv()
			if s, ok := p.setters[v.GetAddress()]; ok {
				s.Set(v.GetValue())
			}
		}
	}
}

func (p *Plugin) observeGetter() {
	for {
		select {
		case n := <-p.vals:
			p.vu.Send(&n)
		case <-p.Ctx.Done():
			return
		}
	}
}
