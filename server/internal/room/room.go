package room

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"net"
	"net/http"
)

const (
	empty = ""
)

type Service interface {
	Join(w http.ResponseWriter, r *http.Request)
}

type service struct {
	rdb   *redis.Client
	rooms map[int]string
}

func NewService(rdb *redis.Client) Service {
	return &service{
		rdb:   rdb,
		rooms: make(map[int]string),
	}
}

type Req struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

func (s *service) Join(w http.ResponseWriter, r *http.Request) {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		http.Error(w, "cannot init http", http.StatusInternalServerError)
		return
	}

	msg, op, err := wsutil.ReadClientData(conn)
	if err != nil {
		http.Error(w, "cannot read the req data", http.StatusInternalServerError)
		return
	}
	var req Req
	err = json.Unmarshal(msg, &req)
	if err != nil {
		http.Error(w, "cannot read the req data", http.StatusInternalServerError)
		return
	}
	if len(req.Username) == 0 {
		http.Error(w, "username is required", http.StatusBadRequest)
		return
	}

	//find some available room
	if len(s.rooms) == 0 {
		//create new room with user id
		s.rooms[0] = req.Username
		b, _ := json.Marshal(&Req{Username: "Server", Message: "waiting people join chat"})
		err = wsutil.WriteServerMessage(conn, op, b)
		if err != nil {
			http.Error(w, "cannot write server message", http.StatusInternalServerError)
			return
		}
		go s.consumeMsg(conn, op, w, req.Username, empty)
		return
	}

	//case found some room

	b, _ := json.Marshal(&Req{Username: "Server", Message: fmt.Sprintf("join %v room!!", s.rooms[0])})
	err = wsutil.WriteServerMessage(conn, op, b)
	if err != nil {
		http.Error(w, "cannot write server message", http.StatusInternalServerError)
		return
	}

	go s.consumeMsg(conn, op, w, req.Username, s.rooms[0])
	delete(s.rooms, 0)
}

func (s *service) consumeMsg(conn net.Conn, op ws.OpCode, w http.ResponseWriter, username, to string) {
	//find the room
	sub := s.rdb.Subscribe(context.Background(), username)
	defer sub.Close()
	targetUser := to
	if len(to) > 0 {
		go s.produceMsg(conn, w, username, targetUser)
	}
	for {
		m, err := sub.ReceiveMessage(context.Background())
		if err != nil {
			http.Error(w, "cannot consume message", http.StatusInternalServerError)
			return
		}

		var reqMsg Req
		if len(targetUser) == 0 {
			err = json.Unmarshal([]byte(m.Payload), &reqMsg)
			if err != nil {
				http.Error(w, "invalid request payload", http.StatusBadRequest)
				return
			}
			targetUser = reqMsg.Username
			go s.produceMsg(conn, w, username, targetUser)
		}

		err = wsutil.WriteServerMessage(conn, op, []byte(m.Payload))
		if err != nil {
			http.Error(w, "cannot write server message", http.StatusInternalServerError)
			return
		}
	}
}

func (s *service) produceMsg(conn net.Conn, w http.ResponseWriter, from, to string) {
	for {
		msg, _, err := wsutil.ReadClientData(conn)
		if err != nil {
			http.Error(w, "cannot update http", http.StatusInternalServerError)
			return
		}
		var req Req
		err = json.Unmarshal(msg, &req)
		if err != nil {
			http.Error(w, "invalid request message", http.StatusInternalServerError)
			return
		}
		fmt.Printf("from \"%v\" to \"%v\" msg: %v\n", from, to, req.Message)

		pb := s.rdb.Publish(context.Background(), to, msg)
		if pb.Err() != nil {
			http.Error(w, fmt.Sprintf("cannot send message to %v", to), http.StatusInternalServerError)
			return
		}
	}
}
