package hub

import (
	"encoding/json"
	"fmt"
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
			logger.Info("Open connection with new rider. user_id: %v, uuid: %v", rider.UserId, rider.UuId)
		case rider := <-h.RiderUnregister:
			fmt.Println(123123123)
			if _, ok := h.Riders[rider]; ok {
				h.Broadcasts = append(h.Broadcasts, Broadcast{
					Rider:     rider,
					Type:      "LEFT",
					Timestamp: time.Now(),
				})
				fmt.Println(123123123)
				delete(h.Riders, rider)
				close(rider.Send)
				fmt.Println(len(h.Riders))
				rider.Hub.Media.StopChannel(rider.MediaKey, rider.ChannelId)
				logger.Info("Closed connection with rider: %v", rider)
			}
		case observer := <-h.ObserverRegister:
			h.Observers[observer] = true

			var users []NotificationUser
			for rider := range h.Riders {
				users = append(users, NotificationUser{
					Id:         "user1",
					PssId:      "",
					Longtitude: rider.Location.Longitude,
					Latitude:   rider.Location.Latitude,
					Videos: NotificationVideos{
						Small:  rider.Endpoint + "/LiveA/live_480272p30_h264.m3u8",
						Medium: rider.Endpoint + "/LiveA/live_720480p30_h264.m3u8",
						Large:  rider.Endpoint + "/LiveA/live_1280x720p60_h264.m3u8",
					},
					Timestamp: time.Now(),
				})
			}
			jsonBytes, _ := json.Marshal(&Notification{
				Type:  "INIT",
				Users: users,
			})
			observer.Send <- jsonBytes
			logger.Info("Open connection with new observer. user_id: %v, uuid: %v", observer.UserId, observer.UuId)
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
				// jsonBytes, err := json.Marshal(h.Broadcasts)
				// if err != nil {
				// 	logger.Error("Failed to marshal broadcasts: %v", err)
				// 	return
				// }
				// for observer := range h.Observers {
				// 	select {
				// 	case observer.Send <- jsonBytes:
				// 	default:
				// 		close(observer.Send)
				// 		// delete(h.clients, client)
				// 	}
				// }
				h.Broadcasts = []Broadcast{}
			}
		}
	}
}
