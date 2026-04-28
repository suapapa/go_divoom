package divoom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type dialTypeResult struct {
	ReturnCode    int      `json:"ReturnCode"`
	ReturnMessage string   `json:"ReturnMessage"`
	DialTypeList  []string `json:"DialTypeList"`
}

func DialType() ([]string, error) {
	resp, err := defaultHTTPClient.Post("https://app.divoom-gz.com/Channel/GetDialType", "", nil)
	if err != nil {
		return nil, fmt.Errorf("fail to get dial type: %w", err)
	}

	var ret dialTypeResult
	err = json.NewDecoder(resp.Body).Decode(&ret)
	closeErr := resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("fail to get dial type: %w", err)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("fail to get dial type: %w", closeErr)
	}

	if ret.ReturnCode != 0 {
		return nil, fmt.Errorf("fail to get dial type: %s", ret.ReturnMessage)
	}

	return ret.DialTypeList, nil
}

type dialListResult struct {
	ReturnCode    int    `json:"ReturnCode"`
	ReturnMessage string `json:"ReturnMessage"`
	TotalNum      string `json:"TotalNum"`
	DialList      []Dial `json:"DialList"`
}
type Dial struct {
	ID   int    `json:"ClockId"`
	Name string `json:"Name"`
}

func DialList(dialType string, page int) ([]Dial, int, error) {
	var buf bytes.Buffer
	data := map[string]interface{}{
		"DialType": dialType,
		"Page":     page,
	}

	jEnc := json.NewEncoder(&buf)
	err := jEnc.Encode(&data)
	if err != nil {
		return nil, 0, fmt.Errorf("fail to get dial list: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://app.divoom-gz.com/Channel/GetDialList", &buf)
	if err != nil {
		return nil, 0, fmt.Errorf("fail to get dial list: %w", err)
	}
	resp, err := defaultHTTPClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("fail to get dial list: %w", err)
	}

	var ret dialListResult
	err = json.NewDecoder(resp.Body).Decode(&ret)
	closeErr := resp.Body.Close()
	if err != nil {
		return nil, 0, fmt.Errorf("fail to get dial list: %w", err)
	}
	if closeErr != nil {
		return nil, 0, fmt.Errorf("fail to get dial list: %w", closeErr)
	}

	if ret.ReturnCode != 0 {
		return nil, 0, fmt.Errorf("fail to get dial list: %s", ret.ReturnMessage)
	}
	tot, _ := strconv.Atoi(ret.TotalNum)
	return ret.DialList, tot, nil
}

func (c *Client) SelectFacesChannel(id int) error {
	err := c.SelectChannel(ChannelFaces)
	if err != nil {
		return fmt.Errorf("fail to select channel to face: %w", err)
	}

	cmd := "Channel/SetClockSelectId"
	data := map[string]interface{}{
		"Command": cmd,
		"ClockId": id,
	}

	return c.doAndCheckErrorCode(data, "fail to select faces channel")
}

type FaceID struct {
	ClockID    int `json:"ClockId"`
	Brightness int `json:"Brightness"`
}

func (c *Client) GetSelectFaceID() (*FaceID, error) {
	cmd := "Channel/GetClockInfo"
	data := map[string]interface{}{
		"Command": cmd,
	}

	resp, err := c.do(data)
	if err != nil {
		return nil, fmt.Errorf("fail to get select face id: %w", err)
	}

	var ret FaceID
	err = json.NewDecoder(resp.Body).Decode(&ret)
	closeErr := resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("fail to get select face id: %w", err)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("fail to get select face id: %w", closeErr)
	}

	return &ret, nil
}
