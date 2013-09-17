package sparql

import (
	"net"
	"net/http"
	"time"
)

type timeoutConfig struct {
	ConnectTimeout   time.Duration
	ReadWriteTimeout time.Duration
}

func timeoutDialer(config *timeoutConfig) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, config.ConnectTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(config.ReadWriteTimeout))
		return conn, nil
	}
}

func newTimeoutClient(open, read time.Duration) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Dial: timeoutDialer(&timeoutConfig{open, read}),
		},
	}
}
