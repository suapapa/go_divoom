package divoom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

type dialListResult struct {
	ReturnCode    int    `json:"ReturnCode"`
	ReturnMessage string `json:"ReturnMessage"`
	TotalNum      int    `json:"TotalNum"`
	DialList      []Dial `json:"DialList"`
}
type Dial struct {
	ID   int    `json:"ClockId"`
	Name string `json:"Name"`
}

func DialList(dialType string, page int) ([]Dial, error) {
	var buf bytes.Buffer
	data := map[string]interface{}{
		"DialType": dialType,
		"Page":     page,
	}

	jEnc := json.NewEncoder(&buf)
	err := jEnc.Encode(&data)
	if err != nil {
		return nil, errors.Wrap(err, "fail to get dial list")
	}

	req, err := http.NewRequest(http.MethodPost, "https://app.divoom-gz.com/Channel/GetDialList", &buf)
	if err != nil {
		return nil, errors.Wrap(err, "fail to get dial list")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "fail to get dial list")
	}
	defer resp.Body.Close()

	var ret dialListResult
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return nil, errors.Wrap(err, "fail to get dial list")
	}

	if ret.ReturnCode != 0 {
		return nil, fmt.Errorf("fail to get dial list: %s", ret.ReturnMessage)
	}

	return ret.DialList, nil
}
