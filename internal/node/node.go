package node

import "time"

type Node struct {
	ID	   string  `json:"id"`
	Name   string  `json:"name"`
	IsMaster bool  `json:"isMaster"`
	IPAddress string `json:"ipAddress"`
	Port int `json:"port"`
	LastSeen time.Time `json:"lastSeen"`
}