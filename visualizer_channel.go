package divoom

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

func (c *Client) VisualizerChannel(idx int) error {
	cmd := "Channel/SetEqPosition"
	data := map[string]interface{}{
		"Command":     cmd,
		"SelectIndex": idx,
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
