package divoom

import (
	"time"
)

func (c *Client) SetCountdownTool(dur time.Duration, start bool) error {
	m := int(dur.Minutes())
	s := int(dur.Seconds()) % 60
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

	return c.doAndCheckErrorCode(data, "fail to set countdown")
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

	return c.doAndCheckErrorCode(data, "fail to set stopwatch")
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

	return c.doAndCheckErrorCode(data, "fail to set scoreboard")
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

	return c.doAndCheckErrorCode(data, "fail to set noise")
}
