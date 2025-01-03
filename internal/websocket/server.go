package websocket

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	ID   string
	Conn *websocket.Conn
	Send chan []byte
}

type Server struct {
	clients    map[string]*Client
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
}

func NewServer() *Server {
	return &Server{
		clients:    make(map[string]*Client),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		mu:         sync.Mutex{},
	}
}

func (s *Server) Run() {
	for {
		select {
		case client := <-s.register:
			s.mu.Lock()
			if existingClient, exists := s.clients[client.ID]; exists {
				existingClient.Conn.Close()
				delete(s.clients, client.ID)
				fmt.Printf("Exisiting client %s disconnected\n", client.ID)
			}
			s.clients[client.ID] = client
			s.mu.Unlock()
			log.Printf("Client %s connected\n", client.ID)

		case client := <-s.unregister:
			if _, ok := s.clients[client.ID]; ok {
				delete(s.clients, client.ID)
				close(client.Send)
				log.Printf("Client %s disconnected\n", client.ID)
			}

		case message := <-s.broadcast:
			for _, client := range s.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(s.clients, client.ID)
				}
			}
		}
	}
}

func (s *Server) HandleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v\n", err)
		return
	}

	clientId := r.RemoteAddr

	client := &Client{
		ID:   clientId,
		Conn: conn,
		Send: make(chan []byte),
	}
	s.register <- client

	go s.handleMessages(client)
	go s.handleWrites(client)
}

func (s *Server) handleMessages(client *Client) {
	defer func() {
		s.unregister <- client
		client.Conn.Close()
	}()

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Unexpected WbeSocket close: %v\n", err)
			}
			break
		}
		s.broadcast <- message
	}
}

func (s *Server) handleWrites(client *Client) {
	for message := range client.Send {
		client.Conn.WriteMessage(websocket.TextMessage, message)
	}
	client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
}

func (s *Server) Shutdown() {
	for _, client := range s.clients {
		close(client.Send)
		client.Conn.Close()
	}
}
