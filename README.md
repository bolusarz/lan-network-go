# LAN Network Go

A Go-based LAN communication system that enables devices to discover each other, elect a master node, and maintain WebSocket connections between nodes.

## Features

- Automatic LAN discovery using UDP broadcast
- Dynamic master node assignment
- WebSocket-based communication between nodes
- No hardcoded IP addresses required
- Support for multiple devices in the network

## Requirements

- Go 1.21 or higher
- Network that allows UDP broadcast (for discovery)

## Installation

```bash
go get github.com/yourusername/lan-network-go
```
