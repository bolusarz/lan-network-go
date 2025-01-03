package discovery

type DiscoveryMessage struct {
	Type string `json:"type"`
	NodeID string `json:"nodeID"`
	IsMaster bool `json:"isMaster"`
	Name string `json:"name"`
	Port int `json:"port"`
}

const (
	MessageTypeRequest = "discoveryRequest"
	MessageTypeResponse = "discoveryResponse"
)