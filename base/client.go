package base

// Client - base client
type Client interface {
	Receive(handler PackageHandler) error
	Send(pkg Package) error
	GetChannelInfo(id string) (ChannelInfo, error)
}

// PackageHandler - handles individual packages
type PackageHandler interface {
	Handle(pkg Package) error
}

// Package - unit of communication between bot and client
type Package struct {
	Author  string
	Channel string
	Message string
}

// ChannelInfo - public channel information
// will retain in cache for a day
type ChannelInfo struct {
	ID           string
	Name         string
	Participants []string
}
