package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int // 菜单模式
}

func NewClient(serverIp string, serverPort int) *Client {
	// 创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}
	// 连接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial err", err)
		return nil
	}

	client.conn = conn

	return client
}

// 用nc命令和net.Dial连接server，消息输出方式不一样
// 处理server返回的消息，标准输出
func (client *Client) DealResponse() {
	// 永久阻塞监听
	io.Copy(os.Stdout, client.conn)
}

// 用户菜单
func (client *Client) menu() bool {
	var flag int

	fmt.Println("1.公聊")
	fmt.Println("2.私聊")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	fmt.Scanln(&flag) // 接收键盘输入

	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println("!!请输入0-3!!")
		return false
	}
}

func (client *Client) SelectUser() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn write err", err)
		return
	}
}

func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	client.SelectUser()
	fmt.Println("请输入私聊对象，exit退出：")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println("请输入聊天内容，exit退出：")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn write err", err)
					break
				}
			}

			chatMsg = ""
			fmt.Println("c请输入聊天内容，exit退出：")
			fmt.Scanln(&chatMsg)
		}

		client.SelectUser()
		fmt.Println("请输入私聊对象，exit退出：")
		fmt.Scanln(&remoteName)
	}

}

func (client *Client) PublicChat() {
	var chatMsg string

	fmt.Println("请输入聊天内容，exit退出：")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		// 发送给server
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn write err", err)
				break
			}
		}

		chatMsg = ""
		fmt.Println("请输入聊天内容，exit退出：")
		fmt.Scanln(&chatMsg)
	}
}

func (client *Client) UpdateName() bool {
	fmt.Println("请输入新用户名：")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err", err)
		return false
	}

	return true
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {
		} // 判断输入

		switch client.flag {
		case 1: // 公聊
			client.PublicChat()
			break
		case 2: // 私聊
			client.PrivateChat()
			break
		case 3: // 更新用户名
			client.UpdateName()
			break
		}
	}
}

var serverIp string
var serverPort int

// 使用flag包进行命令行解析
// ./client -ip 127.0.0.1 -port 8888
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器ip地址，默认127.0.0.1")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口，默认8888")
}

func main() {

	flag.Parse()

	client := NewClient(serverIp, serverPort)

	if client == nil {
		fmt.Println("连接server失败！\n")
		return
	}

	go client.DealResponse()

	fmt.Println("连接server成功！\n")

	client.Run() // 用户菜单

}
