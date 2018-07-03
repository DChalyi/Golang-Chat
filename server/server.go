package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

const proto = "tcp"
const serverMessage = "%s %s"
const userMessage = "[%s %s]: %s\n"
const chatJoinMessage=  "joined chat"
const chatLeaveMessage=  "leaved chat"
const timeFormat  = "2006-01-02 15:04:05"
var serverName string = "Server"



type user struct {
	nickname string
	output   chan message
}

type message struct {
	nickname string
	text     string
}

type safeUsers struct {
	allUsers map[string]user
	mux      sync.Mutex
}

type chat struct {
	users safeUsers
	join  chan user
	leave chan user
	input chan message
}

func usersAcrions(inputChan chan message, username string, messageTempl string, action string)  {
	inputChan <- message{
		serverName,
		fmt.Sprintf(messageTempl, username, action),
	}
}

func (chat *chat) run() {
	for {
		select {

		case user := <-chat.join:
			chat.users.mux.Lock()
			chat.users.allUsers[user.nickname] = user
			go usersAcrions(chat.input, user.nickname, serverMessage, chatJoinMessage)
			chat.users.mux.Unlock()

		case user := <-chat.leave:
			chat.users.mux.Lock()
			delete(chat.users.allUsers, user.nickname)
			go usersAcrions(chat.input, user.nickname, serverMessage, chatLeaveMessage)
			chat.users.mux.Unlock()

		case message := <-chat.input:
			chat.users.mux.Lock()
			for _, user := range chat.users.allUsers {
				select {
				case user.output <- message:
				default:
				}
			}
			chat.users.mux.Unlock()
		}
	}
}

func connectionHandler(connection net.Conn, chat *chat) {

	defer connection.Close()

	io.WriteString(connection, "Enter your nickname: ")
	scanner := bufio.NewScanner(connection)
	scanner.Scan()

	user := user{
		nickname: scanner.Text(),
		output:   make(chan message, 100),
	}

	chat.join <- user
	defer func() {
		chat.leave <- user
	}()

	go func() {
		for scanner.Scan() {
			chat.input <- message{
				user.nickname,
				scanner.Text(),
			}
		}
	}()

	for message := range user.output {
		resp:=fmt.Sprintf(userMessage, message.nickname, time.Now().Format(timeFormat), message.text)
		_, err := io.WriteString(connection, resp)
		if err != nil {
			log.Println(err.Error())
			break
		}
	}
}

func main() {
	server, err := net.Listen(proto, os.Args[1])
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer server.Close()

	chat := &chat{
		safeUsers{allUsers: make(map[string]user)},
		make(chan user),
		make(chan user),
		make(chan message),
	}

	go chat.run()

	for {
		connection, err := server.Accept()
		if err != nil {
			log.Fatalln(err.Error())
		}
		go connectionHandler(connection, chat)
	}
}
