package discovery

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"reach.com/discovery/internal/node"
)

const (
	DiscoveryPort = 8384
	WSPort = 8385
)

type Service struct {
	node *node.Node
	masterFound chan<- string
}

func NewService(node *node.Node, masterFound chan<- string) *Service {
	return &Service{
		node: node,
		masterFound: masterFound,
	}
}

func (s *Service) Start() error {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", DiscoveryPort))
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %w", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on UDP: %w", err)
	}
	defer conn.Close()

	if s.node.IsMaster {
		go s.broadcastMasterPresence()
	}

	return s.handleDiscoveryMessages(conn)
}

func (s *Service) handleDiscoveryMessages(conn *net.UDPConn) error {
	buffer := make([]byte, 1024)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			 fmt.Printf("Error reading UDP message: %v", err)
			 continue
		}

		var msg DiscoveryMessage
		if err := json.Unmarshal(buffer[:n], &msg); err != nil {
			fmt.Printf("Error unmarshaling UDP message: %v", err)
			continue
		}

		switch msg.Type {
			case MessageTypeRequest:
				if s.node.IsMaster {
					response := DiscoveryMessage{
						Type: MessageTypeResponse,
						NodeID: s.node.ID,
						IsMaster: true,
						Port: WSPort,
					}
					responseBytes, _ := json.Marshal(response)
					conn.WriteToUDP(responseBytes, remoteAddr)
				}
			case MessageTypeResponse:
				if msg.IsMaster && !s.node.IsMaster {
					masterAddr := fmt.Sprintf("%s:%d", remoteAddr.IP, msg.Port)
					s.masterFound <- masterAddr
				}
		}
	}
}

func (s *Service) broadcastMasterPresence() {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("255.255.255.255:%d", DiscoveryPort))
	if err != nil {
		fmt.Printf("failed to resolve broadcast address: %v\n", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Printf("failed to create UDP connection: %v\n", err)
		return
	}

	defer conn.Close()

	msg := DiscoveryMessage{
		Type: MessageTypeResponse,
		NodeID: s.node.ID,
		IsMaster: true,
		Port: WSPort,
	}

	msgBytes, _ := json.Marshal(msg)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if _, err := conn.Write(msgBytes); err != nil {
			fmt.Printf("failed to broadcast master presence: %v\n", err)
			return
		}
	}
}