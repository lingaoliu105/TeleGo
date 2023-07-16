package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIP   string
	ServerPort int
	Name       string
	conn       net.Conn
	op         int
}

func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIP:   serverIp,
		ServerPort: serverPort,
		op:         999,
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

func (c *Client) ShowMenu() bool {
	var flg int
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更改用户名")
	fmt.Println("0.退出")
	_, _ = fmt.Scanln(&flg)
	if 0 <= flg && flg < 4 {
		c.op = flg
		return true
	} else {
		fmt.Println("请输入合法范围的数字")
		return false
	}
}

func (c *Client) Run() {
	for c.op != 0 {
		for c.ShowMenu() == false {
			//if operation is invalid, pass
		}
		c.Execute()
	}
}

func (c *Client) UpdateName() bool {
	fmt.Println("enter ur new name: ")
	fmt.Scanln(&c.Name)
	requestMsg := "rename|" + c.Name
	_, err := c.conn.Write([]byte(requestMsg))
	if err != nil {
		fmt.Println(err)
		return false
	} else {
		return true
	}

}

func (c *Client) PublicChat() {
	var message string
	fmt.Println("enter ur public message, enter 'exit' to exit: ")
	_, err := fmt.Scanln(&message)
	if err != nil {
		fmt.Println(err)
	}
	for message != "exit" {
		if len(message) > 0 {
			_, err2 := c.conn.Write([]byte(message))
			if err2 != nil {
				fmt.Println(err2)
			}
		}
		message = ""
		fmt.Println("enter ur public message, enter 'exit' to exit: ")
		fmt.Scanln(&message)
	}
}

func (c *Client) SelectUser() string {
	var name string
	c.conn.Write([]byte("who"))
	fmt.Println("please select a user: ")
	fmt.Scanln(&name)
	return name
}

func (c *Client) PrivateChat() {
	var message string
	var receiver = c.SelectUser()
	fmt.Println("enter ur message: ")
	_, err2 := fmt.Scanln(&message)
	if err2 != nil {
		fmt.Println(err2)
	}
	for message != "exit" {
		_, err3 := c.conn.Write([]byte("@" + receiver + " " + message))
		if err3 != nil {
			fmt.Println(err3)

		}
		message = ""
		fmt.Println("enter ur message: ")
		fmt.Scanln(&message)
	}
}

func (c *Client) Execute() {
	switch c.op {
	case 1: //公聊
		c.PublicChat()
	case 2: //私聊
		c.PrivateChat()
	case 3: //改名
		c.UpdateName()
	}
}

func (c *Client) ProcessResponse() {
	//copy to stdout whenever conn receives, always blocks
	_, err := io.Copy(os.Stdout, c.conn)
	if err != nil {
		return
	}

	//等价于：
	//for {
	//	buf := make([]byte, 4096)
	//	c.conn.Read(buf)
	//	fmt.Println(string(buf))
	//}
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
	go client.ProcessResponse()
	client.Run()
}
