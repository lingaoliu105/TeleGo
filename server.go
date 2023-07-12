package main

import (
	"fmt"
	"net"
)

type Server struct {
	Ip   string
	Port int
}

// create a server
func NewServer(ip string, port int) *Server {
	serverPtr := &Server{
		Ip:   ip,
		Port: port, //note struct creation syntax
	}
	return serverPtr
}

func (s *Server) Handle(conn net.Conn) {
	fmt.Println("connection established")
}

// start server
func (s *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("error creating listener:", err)
		return
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			fmt.Println("error closing listener: ", err)
		}
	}(listener) //defer closing in case forget

	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("error accepting:", err)
			continue
		}

		//do handler
		go s.Handle(conn)
	}

	//close socket
}
