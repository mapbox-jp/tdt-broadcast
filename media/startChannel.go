package media

import (
	"github.com/aws/aws-sdk-go/service/medialive"
)

func (m *Media) StartChannel(userId string, mediaKey string, channelId string) error {
	var err error
	ml := medialive.New(m.Sess)
	if _, err = ml.StartChannel(&medialive.StartChannelInput{
		ChannelId: &channelId,
	}); err != nil {
		return err
	}
	if err = ml.WaitUntilChannelRunning(&medialive.DescribeChannelInput{
		ChannelId: &channelId,
	}); err != nil {
		return err
	}
	if err = m.UpdateStartedChannelOnRedis(mediaKey, userId); err != nil {
		return err
	}
	return nil
}
