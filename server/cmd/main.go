package main

import (
	"fmt"
	"github.com/rkritchat/random-chat-room/server/internal/config"
	"github.com/rkritchat/random-chat-room/server/internal/room"
	"github.com/rkritchat/random-chat-room/server/internal/router"
	"net/http"
)

func main() {
	//init config
	cfg := config.InitConfig()
	defer cfg.Free()

	service := room.NewService(cfg.Rdb)

	//init router
	r := router.InitRouter(service)

	fmt.Println("start on port 9001")
	err := http.ListenAndServe(":9001", r)
	if err != nil {
		panic(err)
	}
}
