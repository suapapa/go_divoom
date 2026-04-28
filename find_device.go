package divoom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
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
	resp, err := defaultHTTPClient.Post("https://app.divoom-gz.com/Device/ReturnSameLANDevice", "", nil)
	if err != nil {
		return nil, fmt.Errorf("fail to find device: %w", err)
	}

	var ret findResult
	err = json.NewDecoder(resp.Body).Decode(&ret)
	closeErr := resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("fail to find device: %w", err)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("fail to find device: %w", closeErr)
	}

	if ret.ReturnCode != 0 {
		return nil, fmt.Errorf("fail to find device: %s", ret.ReturnMessage)
	}

	return ret.DeviceList, nil
}

type Client struct {
	dev        *Device
	url        string
	httpClient *http.Client
}

var defaultHTTPClient = &http.Client{Timeout: 10 * time.Second}

func NewClient(d *Device) *Client {
	return &Client{
		dev:        d,
		url:        fmt.Sprintf("http://%s:80/post", d.DevicePrivateIP),
		httpClient: defaultHTTPClient,
	}
}

func (c *Client) do(data map[string]interface{}) (*http.Response, error) {
	var buf bytes.Buffer
	jEnc := json.NewEncoder(&buf)
	err := jEnc.Encode(&data)
	if err != nil {
		return nil, fmt.Errorf("fail to do: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.url, &buf)
	if err != nil {
		return nil, fmt.Errorf("fail to do: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return c.httpClient.Do(req)
}
