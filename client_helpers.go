package divoom

import (
	"encoding/json"
	"fmt"
)

type errorCode struct {
	ErrorCode int `json:"error_code"`
}

func (c *Client) doAndDecode(data map[string]interface{}, out interface{}, op string) error {
	resp, err := c.do(data)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	decodeErr := json.NewDecoder(resp.Body).Decode(out)
	closeErr := resp.Body.Close()
	if decodeErr != nil {
		return fmt.Errorf("%s: %w", op, decodeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("%s: %w", op, closeErr)
	}

	return nil
}

func (c *Client) doAndCheckErrorCode(data map[string]interface{}, op string) error {
	var ret errorCode
	if err := c.doAndDecode(data, &ret, op); err != nil {
		return err
	}
	if ret.ErrorCode != 0 {
		return fmt.Errorf("%s: %d", op, ret.ErrorCode)
	}
	return nil
}
