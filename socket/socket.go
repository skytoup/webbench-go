package socket

import (
	"fmt"
	"net"
)

const netType = "tcp4"

// 连接socket，连接失败返回nil
func Socket(host *string, port int) *net.TCPConn {
	addr := fmt.Sprintf("%s:%d", *host, port)
	tcpAddr, err := net.ResolveTCPAddr(netType, addr)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	conn, err := net.DialTCP(netType, nil, tcpAddr)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	// conn.
	return conn
}
