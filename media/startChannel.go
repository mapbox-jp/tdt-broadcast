package media

import (
	"github.com/aws/aws-sdk-go/service/medialive"
)

func (m *Media) StartChannel(channelId string) error {
	ml := medialive.New(m.Sess)
	ml.StartChannel(&medialive.StartChannelInput{
		ChannelId: &channelId,
	})
	ml.WaitUntilChannelRunning(&medialive.DescribeChannelInput{
		ChannelId: &channelId,
	})
	return nil
}
