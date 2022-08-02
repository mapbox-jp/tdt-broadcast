package media

import (
	"encoding/json"
	"gps_logger/logger"
)

func (m *Media) SetChannelOnRedis(channelId string, channelKey string, url string) error {
	channel := &Channel{
		Id:     channelKey,
		Url:    url,
		IsUsed: false,
		UserId: "",
	}
	serialized, err := json.Marshal(channel)
	if err != nil {
		logger.Error("Failed to set channels on redis. err: %v", err)
		return err
	}
	m.Rd.HSet("channels", channelId, serialized)
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

func (m *Media) GetChannelFromRedis(channelId string) (*Channel, error) {
	channel := &Channel{}
	value, err := m.Rd.HGet("channels", channelId).Result()
	if err != nil {
		logger.Error("Failed to get channel from redis. channel_id: %v, err: %v", channelId, err)
		return channel, err
	}
	if err := json.Unmarshal([]byte(value), &channel); err != nil {
		logger.Error("Failed unmarshal value. err: %v", err)
		return channel, err
	}
	return channel, nil
}

func (m *Media) UpdateStartedChannelOnRedis(channelId string, userId string) error {
	channel, err := m.GetChannelFromRedis(channelId)
	if err != nil {
		return err
	}
	channel.IsUsed = true
	channel.UserId = userId
	serialized, err := json.Marshal(channel)
	if err != nil {
		logger.Error("Failed to set channels on redis. err: %v", err)
		return err
	}
	m.Rd.HSet("channels", channelId, serialized)
	return nil
}

func (m *Media) UpdateStoppedChannelOnRedis(channelId string) error {
	channel, err := m.GetChannelFromRedis(channelId)
	if err != nil {
		return err
	}
	channel.IsUsed = false
	channel.UserId = ""
	serialized, err := json.Marshal(channel)
	if err != nil {
		logger.Error("Failed to set channels on redis. err: %v", err)
		return err
	}
	m.Rd.HSet("channels", channelId, serialized)
	return nil
}
