package client

import (
	"github.com/roderm/go-adsrv/api/proto/go/address"
	"github.com/roderm/go-adsrv/api/proto/go/registry-service"
	"strings"
)

type notifier struct {
	address string
	vals    chan address.Value
}

type plugin struct {
	cl      registry_service.RegistryClient
	in      registry_service.Registry_InformClient
	vu      registry_service.Registry_ValueUpdateClient
	setters map[string]Settable
	readers map[string]Readable
	vals    chan address.Value
}

func (n *notifier) Value(v []byte) {
	n.vals <- address.Value{
		Address: n.address,
		Value:   v,
	}
}
func NewPlugin(addresses ...IAdress) *plugin {
	ch := make(chan address.Value, 100)
	for _, a := range addresses {
		ao := a.GetObject()
		if ao.Readable {
			r := a.(Readable)
			r.Notifier(&notifier{
				address: ao.Address,
				vals:    ch,
			})
		}
	}
	return nil
}

func (p *plugin) observeSetter() {
	for {
		select {
		default:
			v, _ := p.in.Recv()
			if s, ok := p.setters[v.GetAddress()]; ok {
				s.Set(v.GetValue())
			}
		}
	}
}

func (p *plugin) observeGetter() {
	for {
		select {
		case n := <-p.vals:
			p.vu.Send(&n)
		}
	}
}
