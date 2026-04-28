package divoom

import (
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

	return c.doAndCheckErrorCode(data, "fail to select channel")
}

type CustomIdx int

const (
	CustomIdx0 CustomIdx = iota
	CustomIdx1
	CustomIdx2
)

func (c *Client) CustomChannel(idx CustomIdx) error {
	cmd := "Channel/SetCustomPageIndex"
	data := map[string]interface{}{
		"Command":         cmd,
		"CustomPageIndex": int(idx),
	}

	return c.doAndCheckErrorCode(data, "fail to select channel")
}

func (c *Client) VisualizerChannel(idx int) error {
	err := c.SelectChannel(ChannelVisualizer)
	if err != nil {
		return errors.Wrap(err, "fail to set channel to visualizer")
	}

	cmd := "Channel/SetEqPosition"
	data := map[string]interface{}{
		"Command":    cmd,
		"EqPosition": idx,
	}

	return c.doAndCheckErrorCode(data, "fail to select channel")
}

type CloudChannelIdx int

const (
	CloudChannelRecommendGallery CloudChannelIdx = iota
	CloudChannelFavorite
	CloudChannelSubscribeArtist
)

func (c *Client) CloudChannel(idx CloudChannelIdx) error {
	err := c.SelectChannel(ChannelCloud)
	if err != nil {
		return errors.Wrap(err, "fail to set channel to cloud")
	}

	cmd := "Channel/CloudIndex"
	data := map[string]interface{}{
		"Command": cmd,
		"Index":   idx,
	}

	return c.doAndCheckErrorCode(data, "fail to select channel")
}

type selectIdx struct {
	SelectIdx int `json:"SelectIndex"`
}

func (c *Client) GetCurrentChannel() (Channel, error) {
	cmd := "Channel/GetIndex"
	data := map[string]interface{}{
		"Command": cmd,
	}

	var ret selectIdx
	if err := c.doAndDecode(data, &ret, "fail to get select face id"); err != nil {
		return ChannelInvalid, errors.Wrap(err, "fail to get select face id")
	}

	return Channel(ret.SelectIdx), nil
}
