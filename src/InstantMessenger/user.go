package main

import (
	"net"
	"strings"
)

// user和server其实都是服务端的
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
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,

		server: server,
	}

	go user.ListenMessage()

	return user
}

// 用户上线
func (this *User) Online() {
	// add new online users to the table
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	// broadcast the message that the user is online
	this.server.Broadcast(this, "online")
}

// 用户下线
func (this *User) Offline() {
	// 从onlinemap删除用户
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	// broadcast the message that the user is online
	this.server.Broadcast(this, "offline")
}

// 给当前user的客户端发送消息
func (this *User) sendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

// 用户处理消息
func (this *User) DoMessage(msg string) {

	if msg == "who" { // 查询当前所有在线用户
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线!\n"
			this.sendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" { // 用户重命名

		newName := strings.Split(msg, "|")[1] // 按指定的分隔符将字符串分割成字符串切片

		// 判断name是否已被占用
		_, ok := this.server.OnlineMap[newName]

		if ok { // 查询成功
			this.sendMsg("当前名称已被占用！")
		} else { // name未被占用
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName
			this.sendMsg("用户名已更新：" + this.Name + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" { // 私聊
		remoteName := strings.Split(msg, "|")[1] // 对方用户名
		if remoteName == "" {
			this.sendMsg("消息格式错误，格式：to|name|message")
			return
		}

		remoteUser, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.sendMsg("该用户不存在！\n")
			return
		}

		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.sendMsg("消息内容不能为空！\n")
			return
		}
		// 发送私聊消息
		remoteUser.sendMsg(this.Name + "对您说：" + content)

	} else { // 广播用户的消息
		this.server.Broadcast(this, msg)
	}
}

// monitor user channel
func (this *User) ListenMessage() {
	for {
		msg := <-this.C // read message from user own channel

		// 写入数据到连接
		this.conn.Write([]byte(msg + "\n")) // sending message from user channel to user client
	}
}
