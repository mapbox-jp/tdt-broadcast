package media

import (
	"encoding/json"
	"gps_logger/logger"
)

func (m *Media) SetChannelOnRedis(mediaKey string, channelId string, url string) error {
	channel := &Channel{
		Id:     channelId,
		Url:    url,
		IsUsed: false,
		UserId: "",
	}
	serialized, err := json.Marshal(channel)
	if err != nil {
		logger.Error("Failed to set channels on redis. err: %v", err)
		return err
	}
	m.Rd.HSet("channels", mediaKey, serialized)
	return nil
}

func (m *Media) GetChannelsFromRedis() (map[string]string, error) {
	channels, err := m.Rd.HGetAll("channels").Result()
	if err != nil {
		logger.Error("Failed to get channels from redis. err: %v", err)
		return map[string]string{}, err
	}
	return channels, nil
}

func (m *Media) GetChannelFromRedis(mediaKey string) (*Channel, error) {
	channel := &Channel{}
	value, err := m.Rd.HGet("channels", mediaKey).Result()
	if err != nil {
		logger.Error("Failed getting the channel from redis. media_key: %v, err: %v", mediaKey, err)
		return channel, err
	}
	if err := json.Unmarshal([]byte(value), &channel); err != nil {
		logger.Error("Failed unmarshal value. err: %v", err)
		return channel, err
	}
	return channel, nil
}

func (m *Media) UpdateStartedChannelOnRedis(mediaKey string, userId string) error {
	channel, err := m.GetChannelFromRedis(mediaKey)
	if err != nil {
		return err
	}
	channel.IsUsed = true
	channel.UserId = userId
	serialized, err := json.Marshal(channel)
	if err != nil {
		logger.Error("Failed updating the channel on redis. err: %v", err)
		return err
	}
	m.Rd.HSet("channels", mediaKey, serialized)
	return nil
}

func (m *Media) UpdateStoppedChannelOnRedis(mediaKey string) error {
	channel, err := m.GetChannelFromRedis(mediaKey)
	if err != nil {
		return err
	}
	channel.IsUsed = false
	channel.UserId = ""
	serialized, err := json.Marshal(channel)
	if err != nil {
		logger.Error("Failed updating the channel on redis. err: %v", err)
		return err
	}
	m.Rd.HSet("channels", mediaKey, serialized)
	return nil
}
