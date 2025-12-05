package main

import (
	"bufio"
	"fmt"
	"net/rpc"
	"os"
	"time"
)

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

func main() {
	client, err := rpc.Dial("tcp", "localhost:8081")
	if err != nil {
		fmt.Println("Connection failed:", err)
		return
	}
	defer client.Close()

	var joinReply JoinReply
	err = client.Call("ChatServer.Join", &JoinArgs{}, &joinReply)
	if err != nil {
		fmt.Println("Join failed:", err)
		return
	}

	fmt.Printf("Connected! You are %s\n", joinReply.Name)

	if len(joinReply.History) > 0 {
		fmt.Println("\n--- Chat History ---")
		for _, msg := range joinReply.History {
			fmt.Println(msg)
		}
		fmt.Println("--------------------\n")
	}

	userID := joinReply.UserID

	defer func() {
		client.Call("ChatServer.Leave", &GetUpdatesArgs{UserID: userID}, &SendMessageReply{})
	}()

	go func() {
		for {
			var reply GetUpdatesReply
			err := client.Call("ChatServer.GetUpdates", &GetUpdatesArgs{UserID: userID}, &reply)
			if err != nil {
				return
			}
			for _, msg := range reply.Messages {
				fmt.Println(msg)
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Type your messages:")
	for scanner.Scan() {
		text := scanner.Text()
		if text != "" {
			var reply SendMessageReply
			client.Call("ChatServer.SendMessage", &SendMessageArgs{
				UserID:  userID,
				Message: text,
			}, &reply)
		}
	}
}
