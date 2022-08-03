package hub

import (
	"encoding/json"
	"gps_logger/logger"
	"gps_logger/media"
	"time"

	"github.com/go-redis/redis"
)

const (
	watchPeriod = 3000 * time.Millisecond
)

type Broadcast struct {
	Rider     *Rider `json:"rider"`
	Type      string `json:"type"`
	LiveUrl   string `json:"live_url"`
	Locations Locations
	Timestamp time.Time `json:"timestamp"`
}

type Hub struct {
	Riders             map[*Rider]bool
	Observers          map[*Observer]bool
	Broadcast          chan Broadcast
	Broadcasts         []Broadcast
	RiderRegister      chan *Rider
	RiderUnregister    chan *Rider
	ObserverRegister   chan *Observer
	ObserverUnregister chan *Observer
	Rd                 *redis.Client
	Media              media.MediaRepository
}

func NewHub(rd *redis.Client, media media.MediaRepository) *Hub {
	return &Hub{
		Riders:             make(map[*Rider]bool),
		Observers:          make(map[*Observer]bool),
		Broadcast:          make(chan Broadcast),
		Broadcasts:         []Broadcast{},
		RiderRegister:      make(chan *Rider),
		RiderUnregister:    make(chan *Rider),
		ObserverRegister:   make(chan *Observer),
		ObserverUnregister: make(chan *Observer),
		Rd:                 rd,
		Media:              media,
	}
}

func (h *Hub) Run() {
	ticker := time.NewTicker(watchPeriod)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		case rider := <-h.RiderRegister:
			h.Riders[rider] = true
			h.Broadcasts = append(h.Broadcasts, Broadcast{
				Rider:     rider,
				Type:      "JOINED",
				Locations: Locations{},
				Timestamp: time.Now(),
			})
			jsonBytes, _ := json.Marshal(&Response{
				Type: "JOINED",
			})
			rider.Send <- jsonBytes
			logger.Info("Open connection with new rider: %v", rider)
		case rider := <-h.RiderUnregister:
			if _, ok := h.Riders[rider]; ok {
				h.Broadcasts = append(h.Broadcasts, Broadcast{
					Rider:     rider,
					Type:      "LEFT",
					Timestamp: time.Now(),
				})
				rider.Hub.Media.StopChannel(rider.MediaKey, rider.ChannelId)
				delete(h.Riders, rider)
				close(rider.Send)
				logger.Info("Closed connection with rider: %v", rider)
			}
		case observer := <-h.ObserverRegister:
			h.Observers[observer] = true
			logger.Info("Open connection with new observer: %v", observer)
		case observer := <-h.ObserverUnregister:
			if _, ok := h.Observers[observer]; ok {
				delete(h.Observers, observer)
				close(observer.Send)
				logger.Info("Closed connection with observer: %v", observer)
			}
		case broadcast := <-h.Broadcast:
			h.Broadcasts = append(h.Broadcasts, broadcast)
			userId := broadcast.Rider.UserId
			location := broadcast.Locations[0]
			serialized, _ := json.Marshal(location)
			h.Rd.HSet("locations", userId, serialized)
		case <-ticker.C:
			if len(h.Broadcasts) > 0 {
				jsonBytes, err := json.Marshal(h.Broadcasts)
				if err != nil {
					logger.Error("Failed to marshal broadcasts: %v", err)
					return
				}
				for observer := range h.Observers {
					select {
					case observer.Send <- jsonBytes:
					default:
						close(observer.Send)
						// delete(h.clients, client)
					}
				}
				h.Broadcasts = []Broadcast{}
			}
		}
	}
}
