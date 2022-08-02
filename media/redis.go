package media

import (
	"encoding/json"
	"gps_logger/logger"
)

func (m *Media) SetChannelOnRedis(id string, url string, name string) error {
	channel := &Channel{
		Id:     id,
		Url:    url,
		IsUsed: false,
		UserId: "",
	}
	serialized, err := json.Marshal(channel)
	if err != nil {
		logger.Error("Failed to set channels on redis. err: %v", err)
		return err
	}
	m.Rd.HSet("channels", name, serialized)
	return nil
}

func (m *Media) GetChannelFromRedis() (map[string]string, error) {
	channels, err := m.Rd.HGetAll("channels").Result()
	if err != nil {
		logger.Error("Failed to get channels from redis. err: %v", err)
		return map[string]string{}, err
	}
	return channels, nil
}
