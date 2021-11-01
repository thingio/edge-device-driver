package models

type ProtocolDriver interface {
	Stop(force bool) error
}
