package websocket

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type TelemetryData struct {
	DriverID string  `json:"driver_id"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
}

type DispatchLogic interface {
	AcceptAssignment(ctx context.Context, driverID uuid.UUID, orderID uuid.UUID) error
	RejectAssignment(ctx context.Context, driverID uuid.UUID, orderID uuid.UUID) error
}

type Hub struct {
	clients     map[*Client]bool
	broadcast   chan []byte
	register    chan *Client
	unregister  chan *Client
	redisClient *redis.Client
	svc         DispatchLogic
}

func NewHub(rdb *redis.Client, svc DispatchLogic) *Hub {
	return &Hub{
		broadcast:   make(chan []byte),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		clients:     make(map[*Client]bool),
		redisClient: rdb,
		svc:         svc,
	}
}

func (h *Hub) HandleMessage(client *Client, message []byte) {
	var env Envelope
	if err := json.Unmarshal(message, &env); err != nil {
		log.Printf("invalid json from driver %s: %v", client.driverID, err)
		return
	}

	switch env.Type {
	case MsgLocationUpdate:
		var loc LocationPayload
		if err := json.Unmarshal(env.Payload, &loc); err != nil {
			return
		}

		if err := h.redisClient.GeoAdd(context.Background(), "active_drivers", &redis.GeoLocation{
			Name:      client.driverID,
			Longitude: loc.Lng,
			Latitude:  loc.Lat,
		}).Err(); err != nil {
			log.Printf("failed to update location for driver %s", client.driverID)
		}
		log.Printf("driver %s moved to [%f, %f]", client.driverID, loc.Lat, loc.Lng)
	case MsgOrderResponse:
		var resp OrderResponsePayload
		if err := json.Unmarshal(env.Payload, &resp); err != nil {
			return
		}

		driverUUID, _ := uuid.Parse(client.driverID)
		orderUUID, _ := uuid.Parse(resp.OrderID)

		if resp.Action == "ACCEPT" {
			if err := h.svc.AcceptAssignment(context.Background(), driverUUID, orderUUID); err != nil {
				log.Printf("failed to accept order: %v", err)
				// TODO: send error back to driver
			} else {
				log.Printf("driver %s accepted order %s", client.driverID, resp.OrderID)
			}
		} else if resp.Action == "REJECT" {
			if err := h.svc.RejectAssignment(context.Background(), driverUUID, orderUUID); err != nil {
				log.Printf("failed to reject order: %v", err)
			} else {
				log.Printf("driver %s rejected order %s", client.driverID, resp.OrderID)
			}
		}
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (h *Hub) SendToDriver(driverID string, message any) {
	// TODO: use map[driverID]*Client here for quick lookup
	for client := range h.clients {
		if client.driverID == driverID {
			msgBytes, _ := json.Marshal(message)
			client.send <- msgBytes
			return
		}
	}
}

func (h *Hub) SetService(svc DispatchLogic) {
	h.svc = svc
}
