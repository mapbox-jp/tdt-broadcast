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
	ml.StartChannel(&medialive.StartChannelInput{
		ChannelId: &channelId,
	})
	ml.WaitUntilChannelRunning(&medialive.DescribeChannelInput{
		ChannelId: &channelId,
	})
	return nil
}
