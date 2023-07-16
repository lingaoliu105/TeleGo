package main

import (
	"fmt"
	"io"
	"net"
	"runtime"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	//list of online users
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	//broadcasting channel
	BroadcastC chan string
}

// create a server
func NewServer(ip string, port int) *Server {
	serverPtr := &Server{
		Ip:         ip,
		Port:       port, //note struct creation syntax
		OnlineMap:  make(map[string]*User),
		BroadcastC: make(chan string),
	}
	return serverPtr
}

func (s *Server) Broadcast(user *User, msg string) {
	bcMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	s.BroadcastC <- bcMsg //pass message to broadcasting channel, which is being listened by the func below
}

// listen to broadcasting channel,if there's message, forward to all users
func (s *Server) ListenMessages() {
	for {
		msg := <-s.BroadcastC

		s.mapLock.Lock()
		for _, client := range s.OnlineMap {
			client.C <- msg
		}
		s.mapLock.Unlock()
	}
}

func (s *Server) Handle(conn net.Conn) {
	fmt.Println("connection established")
	user := NewUser(conn, s)
	user.Online()

	//channel monitoring whether user is active
	isAlive := make(chan bool)

	//receive message from client
	go func() {
		buf := make([]byte, 8192)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("read error: ", err)
				return
			}
			msg := string(buf[:n]) //eliminate tailing \n
			if msg[len(msg)-1] == '\n' {
				msg = msg[:len(msg)-1]
			}
			user.ProcessMessage(msg)

			//signal user is alive
			isAlive <- true
		}
	}()

	//current handler blocks
	for {
		select {
		//handle force quitting for inactive users
		case <-isAlive:
			//do nothing
		case <-time.After(time.Second * 600):
			user.Dismiss()
			runtime.Goexit() //or simply return
		}
	}
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

	//start goroutine for listening BroadcastC
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
}
