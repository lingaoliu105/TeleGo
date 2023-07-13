package main

import (
	"fmt"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int

	//list of online users
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	//broadcasting channel
	Message chan string
}

// create a server
func NewServer(ip string, port int) *Server {
	serverPtr := &Server{
		Ip:        ip,
		Port:      port, //note struct creation syntax
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return serverPtr
}

func (s *Server) Broadcast(user *User, msg string) {
	bcMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	s.Message <- bcMsg //pass message to broadcasting channel, which is being listened by the func below
}

// listen to broadcasting channel,if there's message, forward to all users
func (s *Server) ListenMessages() {
	for {
		msg := <-s.Message

		s.mapLock.Lock()
		for _, client := range s.OnlineMap {
			fmt.Println(msg)
			client.C <- msg
		}
		s.mapLock.Unlock()
	}
}

func (s *Server) Handle(conn net.Conn) {
	fmt.Println("connection established")
	user := NewUser(conn)
	//user joined, add to online map
	s.mapLock.Lock()
	s.OnlineMap[user.Name] = user
	s.mapLock.Unlock()

	//broadcast new user's info
	s.Broadcast(user, "已上线\n")

	//receive message from client
	go func() {
		buf := make([]byte, 8192)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				s.Broadcast(user, "下线")
				return
			}
			if err != nil {
				fmt.Println("read error: ", err)
				return
			}
			msg := string(buf[:n-1]) //eliminate tailing \n
			s.Broadcast(user, msg)

		}
	}()

	//current handler blocks

	select {}
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

	//start goroutine for listening Message
	go s.ListenMessages()
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
