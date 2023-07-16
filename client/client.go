package main

import (
	"flag"
	"fmt"
	"net"
)

type Client struct {
	ServerIP   string
	ServerPort int
	Name       string
	conn       net.Conn
}

func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIP:   serverIp,
		ServerPort: serverPort,
	}

	//connect to server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println(err)
		return nil
	}
	client.conn = conn
	return client
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "localhost", "set server's ip address(default localhost)")
	flag.IntVar(&serverPort, "port", 8888, "set server port(default 8888)")
}
func main() {

	flag.Parse() //解析命令行参数

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println("failed starting client")
		return
	}
	fmt.Println("started client")

	select {}
}
