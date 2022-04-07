package divoom

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

type Channel int

const (
	ChannelInvalid Channel = iota - 1
	ChannelFaces
	ChannelCloud
	ChannelVisualizer
	ChannelCustom
)

func (c *Client) SelectChannel(idx Channel) error {
	cmd := "Channel/SetIndex"
	data := map[string]interface{}{
		"Command":     cmd,
		"SelectIndex": int(idx),
	}

	resp, err := c.do(data)
	if err != nil {
		return errors.Wrap(err, "fail to select channel")
	}
	defer resp.Body.Close()

	var ret errorCode
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return errors.Wrap(err, "fail to select channel")
	}

	if ret.ErrorCode != 0 {
		return fmt.Errorf("fail to select channel: %d", ret.ErrorCode)
	}

	return nil
}
