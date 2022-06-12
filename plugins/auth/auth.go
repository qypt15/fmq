package auth

import (
	authfile "github.com/qypt15/fmq/plugins/auth/authfile"
	"github.com/qypt15/fmq/plugins/auth/authhttp"
)

const {
	AuthHTTP = "authhttp"
	AuthFile = "authfile"
}

type Auth interface {
	CheckACL(action,clienID,username,ip,topic string) bool
	CheckConnenct(clientID, username, password sting) bool
}

func NewAuth(name string) Auth {
	switch name {
	case AuthHTTP:
		return authhttp.Init()
	case AuthFile:
		return authfile.Init()
	default:
		return &mockAuth{}
	}
}