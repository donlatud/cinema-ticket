package realtime

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

const MessageTypeSeatUpdate = "SEAT_UPDATE"

type SeatUpdateMessage struct {
	Type   string `json:"type"`
	SeatNo string `json:"seat_no"`
	Status string `json:"status"`
}

type Hub struct {
	mu    sync.RWMutex
	rooms map[string]map[*websocket.Conn]struct{}
}

func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]map[*websocket.Conn]struct{}),
	}
}

func (h *Hub) Register(showtimeID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[showtimeID] == nil {
		h.rooms[showtimeID] = make(map[*websocket.Conn]struct{})
	}
	h.rooms[showtimeID][conn] = struct{}{}
}

func (h *Hub) Unregister(showtimeID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room := h.rooms[showtimeID]
	if room == nil {
		return
	}

	delete(room, conn)
	if len(room) == 0 {
		delete(h.rooms, showtimeID)
	}
}

func (h *Hub) Broadcast(showtimeID string, message SeatUpdateMessage) {
	payload, err := json.Marshal(message)
	if err != nil {
		return
	}

	h.mu.RLock()
	room := h.rooms[showtimeID]
	conns := make([]*websocket.Conn, 0, len(room))
	for conn := range room {
		conns = append(conns, conn)
	}
	h.mu.RUnlock()

	for _, conn := range conns {
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			h.Unregister(showtimeID, conn)
			_ = conn.Close()
		}
	}
}

func (h *Hub) BroadcastSeatUpdate(showtimeID, seatNo, status string) {
	h.Broadcast(showtimeID, SeatUpdateMessage{
		Type:   MessageTypeSeatUpdate,
		SeatNo: seatNo,
		Status: status,
	})
}
