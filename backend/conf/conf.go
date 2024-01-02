package conf

import "buttonmania.win/protocol"

type ContextKey string

const (
	// Context keys for configuration
	KeyConfigPath ContextKey = "configpath"
)

type ClientConf struct {
	ClientId protocol.ClientID `config:"clientId"`
	Rooms    []protocol.RoomID `config:"rooms"`
}

type Conf struct {
	Clients []ClientConf `config:"clients"`
}
