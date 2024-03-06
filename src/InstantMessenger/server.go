package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string // server的地址
	Port int

	OnlineMap map[string]*User // online users table
	mapLock   sync.RWMutex     // lock
	Message   chan string      // message broadcast channel of the server
}

// 创建server的对外接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:   ip,
		Port: port,

		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server
}

// 监听server channel，有消息就发送给所有在线用户
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message // 把消息从server的channel里读出

		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg // 把消息放到user的channel里
		}
		this.mapLock.Unlock()
	}
}

func (this *Server) Broadcast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	this.Message <- sendMsg // 把消息放到server的channel里
}

func (this *Server) Handler(conn net.Conn) {
	// fmt.Println("Connection established successfully")

	user := NewUser(conn, this)

	user.Online()

	isLive := make(chan bool) // 用户是否活跃

	// 接收用户消息并广播
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf) // 从连接里读出数据
			if n == 0 {
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err", err)
				return
			}

			msg := string(buf[:n-1])

			user.DoMessage(msg)

			isLive <- true // 用户消息表示用户活跃
		}
	}()

	// 超时强踢
	for {
		select {
		case <-isLive: // 用户活跃
		case <-time.After(time.Second * 30): // 超时，下线
			// 销毁资源
			user.sendMsg("超时被踢！")
			close(user.C)
			conn.Close()
			// go user.ListenMessage()还未关闭，会造成资源浪费

			return
		}
	}
}

// 启动server
func (this *Server) Start() {
	// socket listen
	// 监听server的地址端口
	// Sprintf生成格式化的字符串，但不打印而是返回
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("Listen err:", err)
		return
	}

	defer listener.Close()

	go this.ListenMessager()

	// 持续处理连接
	for { // endless loop
		conn, err := listener.Accept() // 在服务端使用，用于接受新的连接
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}

		// 处理连接，每来一个用户就创建一个协程
		go this.Handler(conn)
	}

}
