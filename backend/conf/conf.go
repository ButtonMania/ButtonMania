package conf

import "buttonmania.win/protocol"

type ClientConf struct {
	ClientId protocol.ClientID     `config:"clientId"`
	Rooms    []protocol.ButtonType `config:"rooms"`
}

type Conf struct {
	Clients []ClientConf `config:"clients"`
}
