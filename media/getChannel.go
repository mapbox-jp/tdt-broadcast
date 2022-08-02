package media

import (
	"encoding/json"
	"errors"
	"fmt"
	"gps_logger/logger"
)

func (m *Media) GetChannel() (string, string, error) {
	channels, err := m.GetChannelFromRedis()
	if err != nil {
		logger.Error("Failed to get channels from redis. err: %v", err)
		return "", "", err
	}
	for _, value := range channels {
		channel := &Channel{}
		if err := json.Unmarshal([]byte(value), &channel); err != nil {
			logger.Error("Failed unmarshal value. err: %v", err)
			return "", "", err
		}
		fmt.Println(channel)
		if !channel.IsUsed {
			return channel.Id, channel.Url, nil
		}
	}
	logger.Info("There is not unsable channel, all channels are used already.")
	return "", "", errors.New("there is not unsable channel")
}
