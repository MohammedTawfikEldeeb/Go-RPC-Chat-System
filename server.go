package main

import (
	"fmt"
	"net"
	"net/rpc"
	"sync"
)

type ChatServer struct {
	clients   map[int]*Client
	mutex     sync.Mutex
	userCount int
	history   []string
}

type Client struct {
	id   int
	name string
	ch   chan string
}

type JoinArgs struct{}

type JoinReply struct {
	UserID  int
	Name    string
	History []string
}

type SendMessageArgs struct {
	UserID  int
	Message string
}

type SendMessageReply struct {
	Success bool
}

type GetUpdatesArgs struct {
	UserID int
}

type GetUpdatesReply struct {
	Messages []string
}

var chatServer *ChatServer

func (s *ChatServer) Join(args *JoinArgs, reply *JoinReply) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.userCount++
	userID := s.userCount
	username := fmt.Sprintf("User %d", userID)

	client := &Client{
		id:   userID,
		name: username,
		ch:   make(chan string, 100),
	}

	s.clients[userID] = client

	reply.UserID = userID
	reply.Name = username
	reply.History = make([]string, len(s.history))
	copy(reply.History, s.history)

	joinMsg := fmt.Sprintf("%s joined", username)
	s.history = append(s.history, joinMsg)

	for id, c := range s.clients {
		if id != userID {
			select {
			case c.ch <- joinMsg:
			default:
			}
		}
	}

	fmt.Printf("%s connected (ID: %d)\n", username, userID)
	return nil
}

func (s *ChatServer) SendMessage(args *SendMessageArgs, reply *SendMessageReply) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	client, exists := s.clients[args.UserID]
	if !exists {
		reply.Success = false
		return nil
	}

	message := fmt.Sprintf("[%s]: %s", client.name, args.Message)
	s.history = append(s.history, message)

	for id, c := range s.clients {
		if id != args.UserID {
			select {
			case c.ch <- message:
			default:
			}
		}
	}

	reply.Success = true
	return nil
}

func (s *ChatServer) GetUpdates(args *GetUpdatesArgs, reply *GetUpdatesReply) error {
	s.mutex.Lock()
	client, exists := s.clients[args.UserID]
	s.mutex.Unlock()

	if !exists {
		return nil
	}

	reply.Messages = []string{}
	for {
		select {
		case msg := <-client.ch:
			reply.Messages = append(reply.Messages, msg)
		default:
			return nil
		}
	}
}

func (s *ChatServer) Leave(args *GetUpdatesArgs, reply *SendMessageReply) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	client, exists := s.clients[args.UserID]
	if !exists {
		return nil
	}

	leaveMsg := fmt.Sprintf("%s left", client.name)
	s.history = append(s.history, leaveMsg)

	for id, c := range s.clients {
		if id != args.UserID {
			select {
			case c.ch <- leaveMsg:
			default:
			}
		}
	}

	close(client.ch)
	delete(s.clients, args.UserID)

	fmt.Printf("%s disconnected\n", client.name)
	return nil
}

func main() {
	chatServer = &ChatServer{
		clients: make(map[int]*Client),
		history: []string{},
	}

	rpc.Register(chatServer)

	listener, err := net.Listen("tcp", "localhost:8081")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer listener.Close()

	fmt.Println("RPC Chat Server started on localhost:8081")
	fmt.Println("Waiting for clients...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(conn)
	}
}
