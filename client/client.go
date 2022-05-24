package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"net"
	"os"
)

type Msg struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

func main() {
	username := os.Args[1]
	fmt.Printf("hello %v\n", username)
	conn, _, _, err := ws.DefaultDialer.Dial(context.Background(), "ws://127.0.0.1:9001/join")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	b, _ := json.Marshal(&Msg{
		Username: username,
		Message:  "",
	})
	err = wsutil.WriteClientMessage(conn, ws.OpText, b)
	if err != nil {
		panic(err)
	}

	var resp Msg
	msg, _, err := wsutil.ReadServerData(conn)
	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal(msg, &resp)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(resp.Message)

	go consumeMsg(conn)
	for {
		var inputMsg string
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			inputMsg = scanner.Text()
		}

		r, _ := json.Marshal(&Msg{Username: username, Message: inputMsg})
		err = wsutil.WriteClientMessage(conn, ws.OpText, r)
		if err != nil {
			panic(err)
		}
	}
}

func consumeMsg(conn net.Conn) {
	for {
		var req Msg
		msg, _, err := wsutil.ReadServerData(conn)
		if err != nil {
			fmt.Println(err)
			if err.Error() == "EOF" {
				break
			}
			continue
		}

		err = json.Unmarshal(msg, &req)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("%v: %v\n", req.Username, req.Message)
	}
}
