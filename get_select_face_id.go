package divoom

import (
	"encoding/json"

	"github.com/pkg/errors"
)

type FaceID struct {
	ClockID    int `json:"ClockId"`
	Brightness int `json:"Brightness"`
}

func (c *Client) GetSelectFaceID() (*FaceID, error) {
	cmd := "Channel/SetClockInfo"
	data := map[string]interface{}{
		"Command": cmd,
	}

	resp, err := c.do(data)
	if err != nil {
		return nil, errors.Wrap(err, "fail to get select face id")
	}
	defer resp.Body.Close()

	var ret FaceID
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return nil, errors.Wrap(err, "fail to get select face id")
	}

	return &ret, nil
}
