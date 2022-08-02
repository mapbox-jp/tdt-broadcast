package hub

import (
	"bytes"
	"encoding/json"
	"gps_logger/logger"
	"gps_logger/model"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	space = []byte{' '}
)

type Rider struct {
	UuId      string          `json:"-"`
	UserId    string          `json:"-"`
	ChannelId string          `json:"-"`
	Send      chan []byte     `json:"-"`
	Hub       *Hub            `json:"-"`
	Conn      *websocket.Conn `json:"-"`
	Location  model.Location  `json:"-"`
	Timestamp time.Time
}

type Request struct {
	UserId    string    `json:"user_id"`
	Type      string    `json:"type"`
	Locations Locations `json:"locations"`
}
type Response struct {
	Type string
	Url  string
}
type Locations []Location
type Location struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	Timestamp int64   `json:"timestamp"`
}

func (r *Rider) WritePump() {
	ticker := time.NewTicker(pongWait)
	defer func() {
		ticker.Stop()
		r.Conn.Close()
	}()

	for {
		select {
		case text, ok := <-r.Send:
			r.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				r.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := r.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(text)

			n := len(r.Send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-r.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			r.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := r.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (r *Rider) ReadPump() {
	defer func() {
		r.Hub.RiderUnregister <- r
		r.Conn.Close()
	}()

	// r.Conn.SetReadLimit(maxMessageSize)
	// r.Conn.SetReadDeadline(time.Now().Add(pongWait))
	// r.Conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, text, err := r.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("Failed to ws connection: %v", err)
			}
			break
		}
		text = bytes.TrimSpace(bytes.Replace(text, newline, space, -1))
		var req Request
		if err := json.Unmarshal(text, &req); err != nil {
			logger.Error("Failed to unmarshal message: %v", err)
			return
		}

		switch req.Type {
		case "START_BROADCAST":
			channelId, channelKey, url, err := r.Hub.Media.GetChannel()
			if err != nil {
				logger.Error("Failed to set channels on redis. err: %v", err)
			} else {
				r.Hub.Media.StartChannel(r.UserId, channelId)
				r.ChannelId = channelId
			}
			jsonBytes, err := json.Marshal(&Response{
				Type: "START_BROADCAST",
				Url:  url,
			})
			if err != nil {
				logger.Error("Failed to marshal broadcasts: %v", err)
				return
			}
			r.Send <- jsonBytes
		case "PING":
			logger.Debug("Getting ping. userId: %v", r.UserId)
		case "UPDATE_LOCATION":
			r.Hub.Broadcast <- Broadcast{
				Rider:     r,
				Type:      "UPDATE_LOCATION",
				Locations: req.Locations,
				Timestamp: time.Now(),
			}
		case "END":
		}
	}
}

func RiderWorker(userId string, hub *Hub, conn *websocket.Conn, c *gin.Context) {
	rider := &Rider{
		UuId:   c.ClientIP(),
		UserId: userId,
		Send:   make(chan []byte),
		Hub:    hub,
		Conn:   conn,
		Location: model.Location{
			Longitude: 0.0,
			Latitude:  0.0,
			Timestamp: time.Now(),
		},
		Timestamp: time.Now(),
	}
	hub.RiderRegister <- rider

	go rider.WritePump()
	go rider.ReadPump()
}
