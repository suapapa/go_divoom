package divoom

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

var (
	ErrInvalidBrightness   = fmt.Errorf("brightness should be in range of 0~100")
	ErrInvalidWhiteBalance = fmt.Errorf("white balance should be in range of 0~100")
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

	resp, err := c.do(data)
	if err != nil {
		return errors.Wrap(err, "fail to set brightness")
	}
	defer resp.Body.Close()

	var ret errorCode
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return errors.Wrap(err, "fail to set brightness")
	}

	if ret.ErrorCode != 0 {
		return fmt.Errorf("fail to set brightness: %d", ret.ErrorCode)
	}

	return nil
}

func (c *Client) GetAllSetting() (map[string]interface{}, error) {
	cmd := "Channel/GetAllConf"
	data := map[string]interface{}{
		"Command": cmd,
	}

	resp, err := c.do(data)
	if err != nil {
		return nil, errors.Wrap(err, "fail to get all setting")
	}
	defer resp.Body.Close()

	// TODO: make it to struct
	ret := make(map[string]interface{})

	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return nil, errors.Wrap(err, "fail to get all setting")
	}

	if ret["error_code"] != 0 {
		return nil, fmt.Errorf("fail to get all setting: %d", ret["error_code"])
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

	resp, err := c.do(data)
	if err != nil {
		return errors.Wrap(err, "fail to set weather area")
	}
	defer resp.Body.Close()

	var ret errorCode
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return errors.Wrap(err, "fail to set weather area")
	}

	if ret.ErrorCode != 0 {
		return fmt.Errorf("fail to set weather area: %d", ret.ErrorCode)
	}

	return nil
}

func (c *Client) SetTimeZone(timezone string) error {
	cmd := "Sys/TimeZone"
	data := map[string]interface{}{
		"Command":       cmd,
		"TimeZoneValue": timezone,
	}

	resp, err := c.do(data)
	if err != nil {
		return errors.Wrap(err, "fail to set time zone")
	}
	defer resp.Body.Close()

	var ret errorCode
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return errors.Wrap(err, "fail to set time zone")
	}

	if ret.ErrorCode != 0 {
		return fmt.Errorf("fail to set time zone: %d", ret.ErrorCode)
	}

	return nil
}

func (c *Client) SystemTime(utcTime string) error {
	cmd := "Device/SetUTC"
	data := map[string]interface{}{
		"Command": cmd,
		"Utc":     utcTime,
	}

	resp, err := c.do(data)
	if err != nil {
		return errors.Wrap(err, "fail to set utc time")
	}
	defer resp.Body.Close()

	var ret errorCode
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return errors.Wrap(err, "fail to set utc time")
	}

	if ret.ErrorCode != 0 {
		return fmt.Errorf("fail to set utc time: %d", ret.ErrorCode)
	}

	return nil
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

	resp, err := c.do(data)
	if err != nil {
		return errors.Wrap(err, "fail to switch screen")
	}
	defer resp.Body.Close()

	var ret errorCode
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return errors.Wrap(err, "fail to switch screen")
	}

	if ret.ErrorCode != 0 {
		return fmt.Errorf("fail to switch screen: %d", ret.ErrorCode)
	}

	return nil
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

	resp, err := c.do(data)
	if err != nil {
		return nil, errors.Wrap(err, "fail to get device time")
	}
	defer resp.Body.Close()

	var ret DeviceTimeResult
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
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

	resp, err := c.do(data)
	if err != nil {
		return errors.Wrap(err, "fail to set temp mode")
	}
	defer resp.Body.Close()

	var ret errorCode
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return errors.Wrap(err, "fail to set temp mode")
	}

	if ret.ErrorCode != 0 {
		return fmt.Errorf("fail to set temp mode: %d", ret.ErrorCode)
	}

	return nil
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

	resp, err := c.do(data)
	if err != nil {
		return errors.Wrap(err, "fail to set rotation angle")
	}
	defer resp.Body.Close()

	var ret errorCode
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return errors.Wrap(err, "fail to set rotation angle")
	}

	if ret.ErrorCode != 0 {
		return fmt.Errorf("fail to set ratation angle: %d", ret.ErrorCode)
	}

	return nil
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

	resp, err := c.do(data)
	if err != nil {
		return errors.Wrap(err, "fail to set mirror mode")
	}
	defer resp.Body.Close()

	var ret errorCode
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return errors.Wrap(err, "fail to set mirror mode")
	}

	if ret.ErrorCode != 0 {
		return fmt.Errorf("fail to set mirror mode: %d", ret.ErrorCode)
	}

	return nil
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

	resp, err := c.do(data)
	if err != nil {
		return errors.Wrap(err, "fail to set hour mode")
	}
	defer resp.Body.Close()

	var ret errorCode
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return errors.Wrap(err, "fail to set hour mode")
	}

	if ret.ErrorCode != 0 {
		return fmt.Errorf("fail to set hour mode: %d", ret.ErrorCode)
	}

	return nil
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

	resp, err := c.do(data)
	if err != nil {
		return errors.Wrap(err, "fail to set high light mode")
	}
	defer resp.Body.Close()

	var ret errorCode
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return errors.Wrap(err, "fail to set high light mode")
	}

	if ret.ErrorCode != 0 {
		return fmt.Errorf("fail to set high light mode: %d", ret.ErrorCode)
	}

	return nil
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

	resp, err := c.do(data)
	if err != nil {
		return errors.Wrap(err, "fail to set white balance")
	}
	defer resp.Body.Close()

	var ret errorCode
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return errors.Wrap(err, "fail to set white balance")
	}

	if ret.ErrorCode != 0 {
		return fmt.Errorf("fail to set set white balance: %d", ret.ErrorCode)
	}

	return nil

}
