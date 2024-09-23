package server

import (
	"context"
	_ "embed"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	log "github.com/catalystgo/logger/cli"
	"github.com/escalopa/anon-chat-app/domain"

	"github.com/gorilla/websocket"
)

type storage interface {
	Count() int
}

type room interface {
	Register(client *domain.Client)
	Unregister(client *domain.Client)
	SendMessage(msg domain.Message)
}

//go:embed static/index.html
var indexHTML string

type Handler struct {
	httpServer *http.Server
	storage    storage
	room       room
}

func New(port string, storage storage, room room) *Handler {
	h := &Handler{
		httpServer: &http.Server{
			Addr: ":" + port,
		},
		storage: storage,
		room:    room,
	}

	h.httpServer.Handler = h.setupRoutes()
	return h
}

func (h *Handler) Run() error {
	go gracefulShutdown(h.Shutdown)

	log.Infof("Server starting on port %s", h.httpServer.Addr)
	log.Infof("Open http://%s%s in your browser", getAddress(), h.httpServer.Addr)
	if err := h.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (h *Handler) Shutdown(ctx context.Context) error {
	return h.httpServer.Shutdown(ctx)
}

func (h *Handler) setupRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/live", wsHandler(h.room))
	mux.HandleFunc("/count", countHandler(h.storage))
	return mux
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	http.ServeContent(w, r, "index.html", time.Now(), strings.NewReader(indexHTML))
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func wsHandler(room room) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Errorf("Upgrade: %v", err)
			return
		}

		processConn(conn, room)
	}
}

func processConn(conn *websocket.Conn, room room) {
	client := &domain.Client{Conn: conn}
	room.Register(client)

	go func() {
		defer func() {
			room.Unregister(client)
		}()

		for {
			var msg domain.Message
			err := conn.ReadJSON(&msg)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Errorf("Error: %v", err)
				}
				return
			}
			room.SendMessage(msg)
		}
	}()
}

func countHandler(storage storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		count := map[string]int{"count": storage.Count()}
		_ = json.NewEncoder(w).Encode(count) // error ignored for simplicity (it shouldn't occur anyway)
	}
}
