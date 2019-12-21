package client

import (
	pb "github.com/roderm/go-adsrv/api/proto/go/address"
)

type Readable interface {
	IAdress
	Notifier(INotifier)
}
type Settable interface {
	IAdress
	Set(...[]byte) error
}
type INotifier interface {
	Value([]byte)
}
type IAdress interface {
	GetObject() *pb.AddressObject
}
