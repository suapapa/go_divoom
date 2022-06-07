package divoom

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

type PlayGIFType int

const (
	PlayGIFTypeFile PlayGIFType = iota
	PlayGIFTypeFolder
	PlayGIFTypeNet
)

func (c *Client) PlayGif(t PlayGIFType, name string) error {
	cmd := "Device/PlayTFGif"
	data := map[string]interface{}{
		"Command":  cmd,
		"FileType": int(t),
		"FileName": name,
	}

	resp, err := c.do(data)
	if err != nil {
		return errors.Wrap(err, "fail to play gif")
	}
	defer resp.Body.Close()

	var ret errorCode
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return errors.Wrap(err, "fail to play gif")
	}

	if ret.ErrorCode != 0 {
		return fmt.Errorf("fail to play gif: %d", ret.ErrorCode)
	}

	return nil
}

func (c *Client) GetSendingAnimationPicID() (int, error) {
	cmd := "Draw/GetHttpGifId"
	data := map[string]interface{}{
		"Command": cmd,
	}

	resp, err := c.do(data)
	if err != nil {
		return -1, errors.Wrap(err, "fail to get sending animation pic id")
	}
	defer resp.Body.Close()

	ret := make(map[string]interface{})
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return -1, errors.Wrap(err, "fail to get sending animation pic id")
	}

	if ret["error_code"] != float64(0) {
		return -1, fmt.Errorf("fail to get sending animation pic id: %v", ret["error_code"])
	}

	picID := ret["PicId"].(float64)
	return int(picID), nil
}

func (c *Client) ResetSendingAnimationPicID() error {
	cmd := "Draw/ResetHttpGifId"
	data := map[string]interface{}{
		"Command": cmd,
	}

	resp, err := c.do(data)
	if err != nil {
		return errors.Wrap(err, "fail to get sending animation pic id")
	}
	defer resp.Body.Close()

	var ret errorCode
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return errors.Wrap(err, "fail to get sending animation pic id")
	}

	if ret.ErrorCode != 0 {
		return fmt.Errorf("fail to get sending animation pic id: %d", ret.ErrorCode)
	}

	return nil
}

func (c *Client) SendAnimation(width, id, speed int, picDatas [][]byte) error {
	picNum := len(picDatas)
	if picNum > 60 || picNum < 0 {
		return ErrInvalidPicNum
	}

	if width != 64 && width != 32 && width != 16 {
		return ErrInvalidPicWidth
	}

	cmd := "Draw/SendHttpGif"
	for offset := 0; offset < picNum; offset++ {
		picData := picDatas[offset]

		data := map[string]interface{}{
			"Command":   cmd,
			"PicNum":    picNum,
			"PicWidth":  width,
			"PicOffset": offset,
			"PicID":     id,
			"PicSpeed":  speed,
			"PicData":   base64.StdEncoding.EncodeToString(picData),
		}
		resp, err := c.do(data)
		if err != nil {
			return errors.Wrap(err, "fail to send animation")
		}
		defer resp.Body.Close()

		var ret errorCode
		err = json.NewDecoder(resp.Body).Decode(&ret)
		if err != nil {
			return errors.Wrap(err, "fail to send animation")
		}

		if ret.ErrorCode != 0 {
			return fmt.Errorf("fail to send animation: %d", ret.ErrorCode)
		}
	}

	return nil
}

type TextDir int

const (
	TextDirLeft TextDir = iota
	TextDirRight
)

type TextFont int

const (
	TextFont0 TextFont = iota
	TextFont1
	TextFont2
	TextFont3
	TextFont4
	TextFont5
	TextFont6
	TextFont7
)

type TextAlign int

const (
	TextAlignLeft TextAlign = iota + 1
	TextAlignMiddle
	TextAlighRight
)

func (c *Client) SendText(id, x, y int, dir TextDir, font TextFont, width int, str string, speed int, color string, align TextAlign) error {
	cmd := "Draw/SendHttpText"
	data := map[string]interface{}{
		"Command":    cmd,
		"TextId":     id,
		"x":          x,
		"y":          y,
		"dir":        int(dir),
		"font":       int(font),
		"TextWidth":  width,
		"speed":      speed,
		"TextString": str,
		"color":      color,
		"align":      int(align),
	}
	resp, err := c.do(data)
	if err != nil {
		return errors.Wrap(err, "fail to send text")
	}
	defer resp.Body.Close()

	var ret errorCode
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return errors.Wrap(err, "fail to send text")
	}

	if ret.ErrorCode != 0 {
		return fmt.Errorf("fail to send text: %d", ret.ErrorCode)
	}

	return nil
}

func (c *Client) ClearAllTextArea() error {
	cmd := "Draw/ClearHttpText"
	data := map[string]interface{}{
		"Command": cmd,
	}
	resp, err := c.do(data)
	if err != nil {
		return errors.Wrap(err, "fail to clear all text area")
	}
	defer resp.Body.Close()

	var ret errorCode
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return errors.Wrap(err, "fail to clear all text area")
	}

	if ret.ErrorCode != 0 {
		return fmt.Errorf("fail to clear all text area: %d", ret.ErrorCode)
	}

	return nil
}

type getFontListResult struct {
	ReturnCode    int     `json:"ReturnCode"`
	ReturnMessage string  `json:"ReturnMessage"`
	FontList      []*Font `json:"FontList"`
}
type Font struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Width   string `json:"width"`
	High    string `json:"high"`
	Charset string `json:"charset"`
	Type    int    `json:"type"`
}

func GetFontList() ([]*Font, error) {
	resp, err := http.Post("https://app.divoom-gz.com/Device/ReturnSameLANDevice", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "fail to get font list")
	}
	defer resp.Body.Close()

	var ret getFontListResult
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return nil, errors.Wrap(err, "fail to get font list")
	}

	if ret.ReturnCode != 0 {
		return nil, fmt.Errorf("fail to get font list: %s", ret.ReturnMessage)
	}

	return ret.FontList, nil
}

// TODO: TBD
func (c *Client) SendDisplayList() error {
	return ErrNotImplemented
}
