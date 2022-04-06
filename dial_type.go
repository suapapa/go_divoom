package divoom

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

type dialTypeResult struct {
	ReturnCode    int      `json:"ReturnCode"`
	ReturnMessage string   `json:"ReturnMessage"`
	DialTypeList  []string `json:"DialTypeList"`
}

func DialType() ([]string, error) {
	resp, err := http.Post("https://app.divoom-gz.com/Channel/GetDialType", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "fail to get dial type")
	}
	defer resp.Body.Close()

	var ret dialTypeResult
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		return nil, errors.Wrap(err, "fail to get dial type")
	}

	if ret.ReturnCode != 0 {
		return nil, fmt.Errorf("fail to get dial type: %s", ret.ReturnMessage)
	}

	return ret.DialTypeList, nil
}
