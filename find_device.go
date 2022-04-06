package divoom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

type findResult struct {
	ReturnCode    int       `json:"ReturnCode"`
	ReturnMessage string    `json:"ReturnMessage"`
	DeviceList    []*Device `json:"DeviceList"`
}
type Device struct {
	DeviceName      string `json:"DeviceName"`
	DeviceID        int    `json:"DeviceId"`
	DevicePrivateIP string `json:"DevicePrivateIP"`
}

func FindDevice() ([]*Device, error) {
	resp, err := http.Post("https://app.divoom-gz.com/Device/ReturnSameLANDevice", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "fail to find device")
	}
	defer resp.Body.Close()

	var ret findResult
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return nil, errors.Wrap(err, "fail to find device")
	}

	if ret.ReturnCode != 0 {
		return nil, fmt.Errorf("fail to find device: %s", ret.ReturnMessage)
	}

	return ret.DeviceList, nil
}

type Client struct {
	dev *Device
	url string
}

func NewClient(d *Device) *Client {
	return &Client{
		dev: d,
		url: fmt.Sprintf("https://%s:80/post", d.DevicePrivateIP),
	}
}

func (c *Client) do(data map[string]interface{}) (*http.Response, error) {
	var buf bytes.Buffer
	jEnc := json.NewEncoder(&buf)
	err := jEnc.Encode(&data)
	if err != nil {
		return nil, errors.Wrap(err, "fail to do")
	}

	req, err := http.NewRequest(http.MethodPost, c.url, &buf)
	if err != nil {
		return nil, errors.Wrap(err, "fail to do")
	}
	return http.DefaultClient.Do(req)
}
