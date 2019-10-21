package client

import (
	pb "github.com/roderm/go-adsrv/api/proto/go/address"
)

type bla IAdress

type Readable interface {
	Notifier(INotifier)
}
type Settable interface {
	Set([]byte) error
}
type INotifier interface {
	Value([]byte)
}
type IAdress interface {
	GetObject() *pb.AddressObject
}
