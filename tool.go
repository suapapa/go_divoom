package divoom

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

func (c *Client) SetCountdownTool(dur time.Duration, start bool) error {
	m := int(dur.Minutes())
	s := int(dur.Seconds()) / 60
	var v int
	if start {
		v = 1
	}

	cmd := "Tools/SetTimer"
	data := map[string]interface{}{
		"Command": cmd,
		"Minute":  m,
		"Second":  s,
		"Status":  v,
	}

	resp, err := c.do(data)
	if err != nil {
		return errors.Wrap(err, "fail to set countdown")
	}
	defer resp.Body.Close()

	var ret errorCode
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return errors.Wrap(err, "fail to set countdown")
	}

	if ret.ErrorCode != 0 {
		return fmt.Errorf("fail to set countdown: %d", ret.ErrorCode)
	}

	return nil
}

type StopwatchStatus int

const (
	StopwatchStatusStop StopwatchStatus = iota
	StopwatchStatusStart
	StopwatchStatusReset
)

func (c *Client) SetStopwatchTool(s StopwatchStatus) error {
	cmd := "Tools/SetStopWatch"
	data := map[string]interface{}{
		"Command": cmd,
		"Status":  int(s),
	}

	resp, err := c.do(data)
	if err != nil {
		return errors.Wrap(err, "fail to set stopwatch")
	}
	defer resp.Body.Close()

	var ret errorCode
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return errors.Wrap(err, "fail to set stopwatch")
	}

	if ret.ErrorCode != 0 {
		return fmt.Errorf("fail to set stopwatch: %d", ret.ErrorCode)
	}

	return nil
}

func (c *Client) SetScoreboardTool(red, blue int) error {
	if (red < 0 || red > 999) || (blue < 0 || blue > 999) {
		return ErrInvalidScore
	}

	cmd := "Tools/SetScoreBoard"
	data := map[string]interface{}{
		"Command":   cmd,
		"BlueScore": blue,
		"RedScore":  red,
	}

	resp, err := c.do(data)
	if err != nil {
		return errors.Wrap(err, "fail to set scoreboard")
	}
	defer resp.Body.Close()

	var ret errorCode
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return errors.Wrap(err, "fail to set scoreboard")
	}

	if ret.ErrorCode != 0 {
		return fmt.Errorf("fail to set scoreboard: %d", ret.ErrorCode)
	}

	return nil
}

func (c *Client) SetNoiseTool(on bool) error {
	var v int
	if on {
		v = 1
	}

	cmd := "Tools/SetNoiseStatus"
	data := map[string]interface{}{
		"Command":     cmd,
		"NoiseStatus": v,
	}

	resp, err := c.do(data)
	if err != nil {
		return errors.Wrap(err, "fail to set noise")
	}
	defer resp.Body.Close()

	var ret errorCode
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return errors.Wrap(err, "fail to set noise")
	}

	if ret.ErrorCode != 0 {
		return fmt.Errorf("fail to set noise: %d", ret.ErrorCode)
	}

	return nil
}
