package room

import (
	"github.com/escalopa/anon-chat-app/domain"
)

type storage interface {
	Add(msg domain.Message)
	GetAll() []domain.Message
}

type Room struct {
	clients    map[*domain.Client]bool
	broadcast  chan domain.Message
	register   chan *domain.Client
	unregister chan *domain.Client
	storage    storage
}

func New(storage storage) *Room {
	return &Room{
		clients:    make(map[*domain.Client]bool),
		broadcast:  make(chan domain.Message),
		register:   make(chan *domain.Client),
		unregister: make(chan *domain.Client),
		storage:    storage,
	}
}

func (s *Room) Run() {
	for {
		select {
		case client := <-s.register:
			s.clients[client] = true
			s.sendAllMessages(client)

		case client := <-s.unregister:
			if _, ok := s.clients[client]; ok {
				s.closeClient(client)
			}

		case msg := <-s.broadcast:
			s.storage.Add(msg)

			for client := range s.clients {
				err := client.Conn.WriteJSON(msg)
				if err != nil {
					s.closeClient(client)
				}
			}
		}
	}
}

func (s *Room) Register(client *domain.Client) {
	s.register <- client
}

func (s *Room) Unregister(client *domain.Client) {
	s.unregister <- client
}

func (s *Room) SendMessage(msg domain.Message) {
	s.broadcast <- msg
}

func (s *Room) sendAllMessages(client *domain.Client) {
	for _, msg := range s.storage.GetAll() {
		err := client.Conn.WriteJSON(msg)
		if err != nil {
			s.closeClient(client)
		}
	}
}

func (s *Room) closeClient(client *domain.Client) {
	delete(s.clients, client)
	client.Conn.Close()
}
