package domain

import (
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type Client struct {
	Conn *websocket.Conn
}
