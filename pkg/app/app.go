package app

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/google/uuid"
	ws "github.com/gorilla/websocket"
	"reach.com/discovery/internal/discovery"
	"reach.com/discovery/internal/node"
	"reach.com/discovery/internal/websocket"
)

type App struct {
	node        *node.Node
	masterAddr  string
	wsServer    *websocket.Server
	masterFound chan string
	stopChan    chan os.Signal
	wg          sync.WaitGroup
	isMaster    bool
}

func NewApp(name string) *App {
	n := &node.Node{
		ID:       uuid.New().String(),
		Name:     name,
		IsMaster: false,
	}

	return &App{
		node:        n,
		masterFound: make(chan string),
		stopChan:    make(chan os.Signal, 1),
	}
}

func (a *App) SetAsMaster() {
	a.node.IsMaster = true
	a.isMaster = true
}

func (a *App) Start() error {
	fmt.Printf("Starting application for node: %s\n", a.node.Name)

	discoveryService := discovery.NewService(a.node, a.masterFound)
	a.wg.Add(1)

	go func() {
		defer a.wg.Done()
		if err := discoveryService.Start(); err != nil {
			log.Fatalf("Error starting discovery service: %v", err)
		}
	}()

	if a.isMaster {
		a.wsServer = websocket.NewServer()
		a.wg.Add(1)
		go func() {
			defer a.wg.Done()
			a.wsServer.Run()
		}()

		http.HandleFunc("/ws", a.wsServer.HandleConnections)
		a.wg.Add(1)
		go func() {
			defer a.wg.Done()
			fmt.Println("WebScoket server started on :8385")
			if err := http.ListenAndServe(":8385", nil); err != nil {
				log.Fatalf("Error starting websocket server: %v", err)
			}
		}()
	}

	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		a.listenForMaster()
	}()

	signal.Notify(a.stopChan, syscall.SIGINT, syscall.SIGTERM)
	<-a.stopChan
	a.Shutdown()

	return nil
}

func (a *App) listenForMaster() {
	for masterAddr := range a.masterFound {
		fmt.Printf("Master found at: %s", masterAddr)
		a.masterAddr = masterAddr

		a.connectToMaster()
	}
}

func (a *App) connectToMaster() {
	for {
		fmt.Printf("Connecting to master at: %s\n", a.masterAddr)
		conn, _, err := ws.DefaultDialer.Dial(fmt.Sprintf("ws://%s/ws", a.masterAddr), nil)
		if err != nil {
			fmt.Printf("Error connecting to master: %v\n", err)
			continue
		}

		a.handleWebSocket(conn)
		return
	}
}

func (a *App) handleWebSocket(conn *ws.Conn) {
	defer conn.Close()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("Error reading message from master: %v", err)
			break
		}

		fmt.Printf("Received message from master: %s", string(msg))
	}
}

func (a *App) Shutdown() {
	fmt.Println("Shutting down application...")
	close(a.masterFound)
	if a.wsServer != nil {
		a.wsServer.Shutdown()
	}
	a.wg.Wait()
	fmt.Println("Application shutdown complete.")
}
