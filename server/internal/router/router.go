package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/rkritchat/random-chat-room/server/internal/room"
)

func InitRouter(service room.Service) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/join", service.Join)
	return r
}
