package unixsocket

import "os"

var (
	SockAddr string
)

func init() {
	if dir, err := os.Getwd(); err == nil {
		SockAddr = dir + "/" + "ota.sock"
	}
}
