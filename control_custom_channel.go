package divoom

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

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
