package media

import (
	"github.com/aws/aws-sdk-go/service/medialive"
)

func (m *Media) StartChannel(userId string, channelId string) error {
	err := m.UpdateStartedChannelOnRedis(channelId, userId)
	if err != nil {
		return err
	}
	ml := medialive.New(m.Sess)
	_, err = ml.StartChannel(&medialive.StartChannelInput{
		ChannelId: &channelId,
	})
	if err != nil {
		return err
	}
	err = ml.WaitUntilChannelRunning(&medialive.DescribeChannelInput{
		ChannelId: &channelId,
	})
	if err != nil {
		return err
	}
	return nil
}
