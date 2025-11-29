package websocket

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

type TelemetryData struct {
	DriverID string  `json:"driver_id"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
}

type Hub struct {
	clients     map[*Client]bool
	broadcast   chan TelemetryData
	register    chan *Client
	unregister  chan *Client
	redisClient *redis.Client
}

func NewHub(rdb *redis.Client) *Hub {
	return &Hub{
		broadcast:   make(chan TelemetryData),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		clients:     make(map[*Client]bool),
		redisClient: rdb,
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
			cmd := h.redisClient.GeoAdd(context.Background(), "active_drivers", &redis.GeoLocation{
				Name:      message.DriverID,
				Longitude: message.Lng,
				Latitude:  message.Lat,
			})
			if err := cmd.Err(); err != nil {
				log.Printf("redis error: %v", err)
			}

			log.Printf("live tracking: driver %s is at [%f, %f]", message.DriverID, message.Lat, message.Lng)
		}
	}
}
