package divoom

import (
	"encoding/json"

	"github.com/pkg/errors"
)

type selectIdx struct {
	SelectIdx int `json:"SelectIndex"`
}

func (c *Client) GetCurrentChannel() (Channel, error) {
	cmd := "Channel/GetIndex"
	data := map[string]interface{}{
		"Command": cmd,
	}

	resp, err := c.do(data)
	if err != nil {
		return ChannelInvalid, errors.Wrap(err, "fail to get select face id")
	}
	defer resp.Body.Close()

	var ret selectIdx
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return ChannelInvalid, errors.Wrap(err, "fail to get select face id")
	}

	return Channel(ret.SelectIdx), nil
}
