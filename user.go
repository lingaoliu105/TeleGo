package main

import (
	"fmt"
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

// create user
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	go user.ListenMessage() //start listening
	return user
}

// messages from server via channel
func (u *User) ListenMessage() {
	for {
		msg := <-u.C
		u.Send(msg)
	}
}

// 用户上线
func (u *User) Online() {
	u.server.mapLock.Lock()
	u.server.OnlineMap[u.Name] = u
	u.server.mapLock.Unlock()

	//broadcast new user's info
	u.server.Broadcast(u, "已上线\n")
}

// 用户下线
func (u *User) Offline() {
	u.server.mapLock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.mapLock.Unlock()

	//broadcast new user's info
	u.server.Broadcast(u, "下线\n")
}

func (u *User) GetLock() {
	u.server.mapLock.Lock()
}

func (u *User) ReleaseLock() {
	u.server.mapLock.Unlock()
}

func (u *User) Send(message string) {
	if message[len(message)-1] != '\n' {
		message = message + "\n"
	}
	_, err := u.conn.Write([]byte(message))
	if err != nil {
		fmt.Println("error writing to client")
		return
	}
}

// process incoming messages from this user
func (u *User) ProcessMessage(msg string) {
	//query online users
	if msg == "who" {
		u.GetLock()
		for _, onlineUser := range u.server.OnlineMap {
			reply := "[" + onlineUser.Addr + "]" + onlineUser.Name + ": 在线\n"
			u.Send(reply)
		}
		u.ReleaseLock()
	} else if len(msg) > 7 && msg[:7] == "rename|" { //rename request
		newName := strings.Split(msg, "|")[1]

		//check if name is available
		u.GetLock()
		_, ok := u.server.OnlineMap[newName]
		if ok {
			u.Send("name is not available")
			u.ReleaseLock()
		} else {
			u.server.OnlineMap[newName] = u
			delete(u.server.OnlineMap, u.Name)
			u.ReleaseLock()
			u.Name = newName
			u.Send("成功更名为：" + newName)
		}
	} else { //normal messages, broadcast to all clients
		u.server.Broadcast(u, msg)
	}
}

func (u *User) Dismiss() {
	close(u.C)
	_, err2 := u.conn.Write([]byte("您已下线\n"))
	if err2 != nil {
		fmt.Println(err2)
		return
	}
	err := u.conn.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	u.GetLock()
	delete(u.server.OnlineMap, u.Name)
	u.ReleaseLock()
}
