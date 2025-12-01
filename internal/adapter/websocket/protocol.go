package websocket

import "encoding/json"

type MessageType string

const (
	MsgLocationUpdate MessageType = "LOCATION_UPDATE"
	MsgOrderResponse  MessageType = "ORDER_RESPONSE"
)

type Envelope struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type OrderResponsePayload struct {
	OrderID string `json:"order_id"`
	Action  string `json:"action"`
}

type LocationPayload struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}
