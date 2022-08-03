package hub

import (
	"gps_logger/logger"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 10 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
)

type Observer struct {
	UuId   string `json:"-"`
	UserId string `json:"-"`
	Send   chan []byte
	Hub    *Hub
	Conn   *websocket.Conn
}

func (o *Observer) WritePump() {
	ticker := time.NewTicker(pongWait)
	defer func() {
		ticker.Stop()
		o.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-o.Send:
			o.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				o.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := o.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				logger.Error("Failed by connection next writer: %v", err)
				return
			}
			w.Write(message)

			n := len(o.Send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-o.Send)
			}

			if err := w.Close(); err != nil {
				logger.Error("Failed by unknown error: %v", err)
				return
			}
		case <-ticker.C:
			o.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := o.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// func (o *Observer) ReadPump() {
// 	defer func() {
// 		o.Hub.Unregister <- o
// 		o.Conn.Close()
// 	}()

// 	// o.Conn.SetReadLimit(maxMessageSize)
// 	// o.Conn.SetReadDeadline(time.Now().Add(pongWait))
// 	// o.Conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

// 	for {
// 		_, message, err := o.Conn.ReadMessage()
// 		if err != nil {
// 			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
// 				logger.Error("Failed to ws connection: %v", err)
// 			}
// 			break
// 		}
// 		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
// 		// o.Hub.Broadcast <- message
// 	}
// }

func ObserverWorker(hub *Hub, conn *websocket.Conn, c *gin.Context) {
	observer := &Observer{
		UuId: "",
		Send: make(chan []byte),
		Hub:  hub,
		Conn: conn,
	}
	hub.ObserverRegister <- observer

	go observer.WritePump()
	// go observer.ReadPump()
}
