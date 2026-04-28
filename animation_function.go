package divoom

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/gif"
	"os"

	"github.com/nfnt/resize"
	"github.com/oliamb/cutter"
)

type PlayGifType int

const (
	PlayGifTypeFile PlayGifType = iota
	PlayGifTypeFolder
	PlayGifTypeNet
)

func (c *Client) PlayTFGif(t PlayGifType, name string) error {
	cmd := "Device/PlayTFGif"
	data := map[string]interface{}{
		"Command":  cmd,
		"FileType": int(t),
		"FileName": name,
	}

	return c.doAndCheckErrorCode(data, "fail to play gif")
}

type getSendingAnimationPicIDResult struct {
	ErrorCode int `json:"error_code"`
	PicID     int `json:"PicId"`
}

func (c *Client) GetSendingAnimationPicID() (int, error) {
	cmd := "Draw/GetHttpGifId"
	data := map[string]interface{}{
		"Command": cmd,
	}
	resp, err := c.do(data)
	if err != nil {
		return -1, fmt.Errorf("fail to get sending animation pic id: %w", err)
	}

	var ret getSendingAnimationPicIDResult
	err = json.NewDecoder(resp.Body).Decode(&ret)
	closeErr := resp.Body.Close()
	if err != nil {
		return -1, fmt.Errorf("fail to get sending animation pic id: %w", err)
	}
	if closeErr != nil {
		return -1, fmt.Errorf("fail to get sending animation pic id: %w", closeErr)
	}

	if ret.ErrorCode != 0 {
		return -1, fmt.Errorf("fail to get sending animation pic id: %d", ret.ErrorCode)
	}

	return ret.PicID, nil
}

func (c *Client) ResetSendingAnimationPicID() error {
	cmd := "Draw/ResetHttpGifId"
	data := map[string]interface{}{
		"Command": cmd,
	}

	return c.doAndCheckErrorCode(data, "fail to get sending animation pic id")
}

func (c *Client) SendAnimationGif(id int, gifImg *gif.GIF) error {
	frameCnt := len(gifImg.Image)
	imgs := make([]image.Image, frameCnt)
	delayMSecs := make([]int, frameCnt)

	// find short length in w and h
	bound := gifImg.Image[0].Bounds()
	w, h := bound.Dx(), bound.Dy()
	var length int
	if w < h {
		length = int(w)
	} else {
		length = int(h)
	}

	var targetLength uint
	if length >= 64 {
		targetLength = 64
	} else if length >= 32 {
		targetLength = 32
	} else {
		targetLength = 16
	}

	// crop center rect
	cutterCfg := cutter.Config{
		Width:  length,
		Height: length,
		Mode:   cutter.Centered,
	}

	for i := 0; i < frameCnt; i++ {
		cImg, err := cutter.Crop(gifImg.Image[i], cutterCfg)
		if err != nil {
			return fmt.Errorf("fail to set gif: %w", err)
		}

		// resize
		if uint(length) != targetLength {
			imgs[i] = resize.Resize(targetLength, targetLength, cImg, resize.Lanczos3)
		} else {
			imgs[i] = cImg
		}

		delayMSecs[i] = gifImg.Delay[i] * 10
	}

	return c.SendAnimationImgs(id, delayMSecs, imgs)
}

func (c *Client) SendAnimationImgs(id int, speedMSecs []int, imgs []image.Image) error {
	if len(imgs) < 1 {
		return fmt.Errorf("want more than one image")
	}
	w0, h0 := imgs[0].Bounds().Dx(), imgs[0].Bounds().Dy()
	if w0 != h0 {
		return fmt.Errorf("want rectangle image")
	}
	if w0 != 64 && w0 != 32 && w0 != 16 {
		return ErrInvalidPicWidth
	}

	picDatas := make([][]byte, len(imgs))
	for i := range picDatas {
		picDatas[i] = imgToRGB24Bytes(imgs[i])
	}

	return c.SendAnimation(w0, id, speedMSecs, picDatas)
}

func (c *Client) SendAnimation(width, id int, speedMSecs []int, picDatas [][]byte) error {
	picNum := len(picDatas)
	if picNum > 60 || picNum < 0 {
		return ErrInvalidPicNum
	}
	if len(speedMSecs) != picNum {
		return fmt.Errorf("speedMSecs length mismatch: got %d, want %d", len(speedMSecs), picNum)
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
			"PicSpeed":  speedMSecs[offset],
			"PicData":   base64.StdEncoding.EncodeToString(picData),
		}
		resp, err := c.do(data)
		if err != nil {
			return fmt.Errorf("fail to send animation: %w", err)
		}

		var ret errorCode
		err = json.NewDecoder(resp.Body).Decode(&ret)
		closeErr := resp.Body.Close()
		if err != nil {
			return fmt.Errorf("fail to send animation: %w", err)
		}
		if closeErr != nil {
			return fmt.Errorf("fail to send animation: %w", closeErr)
		}

		if ret.ErrorCode != 0 {
			return fmt.Errorf("fail to send animation: %d", ret.ErrorCode)
		}
	}

	return nil
}

func imgToRGB24Bytes(img image.Image) []byte {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	imgData := make([]byte, w*h*3)
	var i int
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			r8, g8, b8 := r>>8, g>>8, b>>8
			imgData[i] = byte(r8)
			imgData[i+1] = byte(g8)
			imgData[i+2] = byte(b8)
			i += 3
		}
	}
	return imgData
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
	return c.doAndCheckErrorCode(data, "fail to send text")
}

func (c *Client) ClearAllTextArea() error {
	cmd := "Draw/ClearHttpText"
	data := map[string]interface{}{
		"Command": cmd,
	}
	return c.doAndCheckErrorCode(data, "fail to clear all text area")
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
	resp, err := defaultHTTPClient.Post("https://app.divoom-gz.com/Device/ReturnSameLANDevice", "", nil)
	if err != nil {
		return nil, fmt.Errorf("fail to get font list: %w", err)
	}

	var ret getFontListResult
	err = json.NewDecoder(resp.Body).Decode(&ret)
	closeErr := resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("fail to get font list: %w", err)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("fail to get font list: %w", closeErr)
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

// --- for mcp

func (c *Client) PlayGifFile(gifPath string) error {
	f, err := os.Open(gifPath)
	if err != nil {
		return fmt.Errorf("fail to play gif file: %w", err)
	}
	defer f.Close()

	gifImg, err := gif.DecodeAll(f)
	if err != nil {
		return fmt.Errorf("fail to send animation gif path: %w", err)
	}

	if err := c.ResetSendingAnimationPicID(); err != nil {
		return fmt.Errorf("fail to reset sending animation pic id: %w", err)
	}

	return c.SendAnimationGif(1, gifImg)
}
