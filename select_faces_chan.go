package divoom

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

type errorCode struct {
	ErrorCode int `json:"error_code"`
}

func (c *Client) SelectFacesChannel(id int) error {
	cmd := "Channel/SetClockSelectId"
	data := map[string]interface{}{
		"Command": cmd,
		"ClockId": id,
	}

	resp, err := c.do(data)
	if err != nil {
		return errors.Wrap(err, "fail to select faces channel")
	}
	defer resp.Body.Close()

	var ret errorCode
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return errors.Wrap(err, "fail to select faces channel")
	}

	if ret.ErrorCode != 0 {
		return fmt.Errorf("fail to select faces channel: %d", ret.ErrorCode)
	}

	return nil
}
