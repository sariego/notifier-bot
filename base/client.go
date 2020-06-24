package base

// Client - base client
// Receive and Send are needed for basic functionality
type Client interface {
	Receive(handler func(pkg Package)) error
	Send(pkg Package) error
	GetChannelInfo(id string) (ChannelInfo, error)
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
