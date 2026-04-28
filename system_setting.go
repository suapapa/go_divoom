package divoom

import (
	"fmt"

	"github.com/pkg/errors"
)

var (
	ErrInvalidBrightness   = fmt.Errorf("brightness should be in range of 0~100")
	ErrInvalidWhiteBalance = fmt.Errorf("white balance should be in range of 0~100")
	ErrInvalidScore        = fmt.Errorf("score should be in range of 0~999")
	ErrInvalidPicNum       = fmt.Errorf("pic num should be smaller than 60")
	ErrInvalidPicWidth     = fmt.Errorf("pic width should be 16, 32 or 64")
	ErrNotImplemented      = fmt.Errorf("not implemented")
)

func (c *Client) SetBrightness(brightness int) error {
	if brightness < 0 || brightness > 100 {
		return ErrInvalidBrightness
	}

	cmd := "Channel/SetBrightness"
	data := map[string]interface{}{
		"Command":    cmd,
		"Brightness": brightness,
	}

	return c.doAndCheckErrorCode(data, "fail to set brightness")
}

func (c *Client) GetAllSetting() (map[string]interface{}, error) {
	cmd := "Channel/GetAllConf"
	data := map[string]interface{}{
		"Command": cmd,
	}

	// TODO: make it to struct
	ret := make(map[string]interface{})
	if err := c.doAndDecode(data, &ret, "fail to get all setting"); err != nil {
		return nil, errors.Wrap(err, "fail to get all setting")
	}
	rawCode, ok := ret["error_code"]
	if !ok {
		return nil, fmt.Errorf("fail to get all setting: missing error_code")
	}
	errorCodeNum, ok := rawCode.(float64)
	if !ok {
		return nil, fmt.Errorf("fail to get all setting: invalid error_code type %T", rawCode)
	}
	if int(errorCodeNum) != 0 {
		return nil, fmt.Errorf("fail to get all setting: %d", int(errorCodeNum))
	}

	return ret, nil
}

func (c *Client) WeatherAreaSetting(long, lat string) error {
	cmd := "Sys/LogAndLat"
	data := map[string]interface{}{
		"Command":  cmd,
		"Logitude": long,
		"Latitude": lat,
	}

	return c.doAndCheckErrorCode(data, "fail to set weather area")
}

func (c *Client) SetTimeZone(timezone string) error {
	cmd := "Sys/TimeZone"
	data := map[string]interface{}{
		"Command":       cmd,
		"TimeZoneValue": timezone,
	}

	return c.doAndCheckErrorCode(data, "fail to set time zone")
}

func (c *Client) SystemTime(utcTime string) error {
	cmd := "Device/SetUTC"
	data := map[string]interface{}{
		"Command": cmd,
		"Utc":     utcTime,
	}

	return c.doAndCheckErrorCode(data, "fail to set utc time")
}

func (c *Client) ScreenSwitch(on bool) error {
	cmd := "Channel/OnOffScreen"
	var val int
	if on {
		val = 1
	}
	data := map[string]interface{}{
		"Command": cmd,
		"OnOff":   val,
	}

	return c.doAndCheckErrorCode(data, "fail to switch screen")
}

type DeviceTimeResult struct {
	ErrorCode int    `json:"error_code"`
	UTCTime   int    `json:"UTCTime"`
	LocalTime string `json:"LocalTime"`
}

func (c *Client) GetDeviceTime() (*DeviceTimeResult, error) {
	cmd := "Device/GetDeviceTime"
	data := map[string]interface{}{
		"Command": cmd,
	}

	var ret DeviceTimeResult
	if err := c.doAndDecode(data, &ret, "fail to get device time"); err != nil {
		return nil, errors.Wrap(err, "fail to get device time")
	}

	if ret.ErrorCode != 0 {
		return nil, fmt.Errorf("fail to get device time: %d", ret.ErrorCode)
	}

	return &ret, nil
}

type TempMode int

const (
	TempModeCelsius TempMode = iota
	TempModeFahrenheit
)

func (c *Client) SetTemperatureMode(tempMode TempMode) error {
	cmd := "Device/SetDisTempMode"
	data := map[string]interface{}{
		"Command": cmd,
		"Mode":    int(tempMode),
	}

	return c.doAndCheckErrorCode(data, "fail to set temp mode")
}

type RotationAngle int

const (
	RotationAngle0 RotationAngle = iota
	RotationAngle90
	RotationAngle180
	RotationAngle270
)

func (c *Client) SetRotationAngle(angle RotationAngle) error {
	cmd := "Device/SetScreenRotationAngle"
	data := map[string]interface{}{
		"Command": cmd,
		"Mode":    int(angle),
	}

	return c.doAndCheckErrorCode(data, "fail to set rotation angle")
}

type MirrorMode int

const (
	MirrorModeDisable MirrorMode = iota
	MirrorModeEnable
)

func (c *Client) SetMirrorMode(on MirrorMode) error {
	cmd := "Device/SetMirrorMode"
	data := map[string]interface{}{
		"Command": cmd,
		"Mode":    int(on),
	}

	return c.doAndCheckErrorCode(data, "fail to set mirror mode")
}

type HourMode int

const (
	HourMode12 HourMode = iota
	HourMode24
)

func (c *Client) SetHourMode(hm HourMode) error {
	cmd := "Device/SetTime24Flag"
	data := map[string]interface{}{
		"Command": cmd,
		"Mode":    int(hm),
	}

	return c.doAndCheckErrorCode(data, "fail to set hour mode")
}

func (c *Client) SetHighLightMode(on bool) error {
	var val int
	if on {
		val = 1
	}
	cmd := "Device/SetHighLightMode"
	data := map[string]interface{}{
		"Command": cmd,
		"Mode":    int(val),
	}

	return c.doAndCheckErrorCode(data, "fail to set high light mode")
}

func (c *Client) SetWhiteBalance(r, g, b int) error {
	if (r < 0 || r > 100) || (g < 0 || g > 100) || (b < 0 || b > 100) {
		return ErrInvalidWhiteBalance
	}

	cmd := "Device/SetHighLightMode"
	data := map[string]interface{}{
		"Command": cmd,
		"RValue":  r,
		"GValue":  g,
		"BValue":  b,
	}

	return c.doAndCheckErrorCode(data, "fail to set white balance")

}
