package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/gif"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"


	divoom "github.com/suapapa/go_divoom"
)

const (
	mcpProtocolVersion = "2024-11-05"
	serverName         = "go-divoom-mcp"
	serverVersion      = "0.1.0"
)

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type toolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type mcpTool struct {
	Name        string
	Description string
	InputSchema map[string]interface{}
	Handler     func(map[string]interface{}) (interface{}, error)
}

type mcpServer struct {
	mu     sync.RWMutex
	device *divoom.Device
	client *divoom.Client
	tools  map[string]mcpTool
}

func newMCPServer() *mcpServer {
	s := &mcpServer{
		tools: make(map[string]mcpTool),
	}
	s.registerTools()
	return s
}

func (s *mcpServer) registerTool(t mcpTool) {
	s.tools[t.Name] = t
}

func (s *mcpServer) registerTools() {
	s.registerTool(mcpTool{
		Name:        "find_device",
		Description: "Find Divoom devices on same LAN.",
		InputSchema: schemaObject(nil, nil),
		Handler: func(_ map[string]interface{}) (interface{}, error) {
			devs, err := divoom.FindDevice()
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{"devices": devs}, nil
		},
	})

	s.registerTool(mcpTool{
		Name:        "connect_device",
		Description: "Connect to a Divoom device by id/ip/index from FindDevice.",
		InputSchema: schemaObject(
			map[string]interface{}{
				"device_id": schemaInteger("Target device id"),
				"ip":        schemaString("Target device private ip"),
				"index":     schemaInteger("Index from find_device result"),
			},
			nil,
		),
		Handler: s.handleConnectDevice,
	})

	s.registerTool(mcpTool{
		Name:        "get_connected_device",
		Description: "Get currently connected device.",
		InputSchema: schemaObject(nil, nil),
		Handler: func(_ map[string]interface{}) (interface{}, error) {
			s.mu.RLock()
			defer s.mu.RUnlock()
			return map[string]interface{}{"device": s.device}, nil
		},
	})

	s.registerTool(mcpTool{
		Name:        "disconnect_device",
		Description: "Disconnect currently connected device.",
		InputSchema: schemaObject(nil, nil),
		Handler: func(_ map[string]interface{}) (interface{}, error) {
			s.mu.Lock()
			s.device = nil
			s.client = nil
			s.mu.Unlock()
			return map[string]interface{}{"ok": true}, nil
		},
	})

	s.registerTool(mcpTool{
		Name:        "dial_type",
		Description: "List available dial types.",
		InputSchema: schemaObject(nil, nil),
		Handler: func(_ map[string]interface{}) (interface{}, error) {
			dialTypes, err := divoom.DialType()
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{"dial_types": dialTypes}, nil
		},
	})

	s.registerTool(mcpTool{
		Name:        "dial_list",
		Description: "List dials by type and page.",
		InputSchema: schemaObject(
			map[string]interface{}{
				"dial_type": schemaString("Dial type from dial_type"),
				"page":      schemaInteger("Page number"),
			},
			[]string{"dial_type", "page"},
		),
		Handler: func(args map[string]interface{}) (interface{}, error) {
			dialType, err := getString(args, "dial_type", true)
			if err != nil {
				return nil, err
			}
			page, err := getInt(args, "page", true)
			if err != nil {
				return nil, err
			}
			dials, total, err := divoom.DialList(dialType, page)
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{
				"total": total,
				"dials": dials,
			}, nil
		},
	})

	s.registerTool(mcpTool{
		Name:        "get_font_list",
		Description: "Get font list supported by Divoom API.",
		InputSchema: schemaObject(nil, nil),
		Handler: func(_ map[string]interface{}) (interface{}, error) {
			fonts, err := divoom.GetFontList()
			if err != nil {
				return nil, err
			}
			return map[string]interface{}{"fonts": fonts}, nil
		},
	})

	s.registerTool(mcpTool{
		Name:        "play_gif",
		Description: "Play a GIF file/folder/net source on device.",
		InputSchema: schemaObject(
			map[string]interface{}{
				"type": schemaInteger("0=file, 1=folder, 2=net"),
				"name": schemaString("Name/path/url"),
			},
			[]string{"type", "name"},
		),
		Handler: func(args map[string]interface{}) (interface{}, error) {
			t, err := getInt(args, "type", true)
			if err != nil {
				return nil, err
			}
			name, err := getString(args, "name", true)
			if err != nil {
				return nil, err
			}
			return s.withClient(func(c *divoom.Client) (interface{}, error) {
				if err := c.PlayGif(divoom.PlayGIFType(t), name); err != nil {
					return nil, err
				}
				return map[string]interface{}{"ok": true}, nil
			})
		},
	})

	s.registerTool(mcpTool{
		Name:        "get_sending_animation_pic_id",
		Description: "Get current HTTP animation pic id.",
		InputSchema: schemaObject(nil, nil),
		Handler: func(_ map[string]interface{}) (interface{}, error) {
			return s.withClient(func(c *divoom.Client) (interface{}, error) {
				id, err := c.GetSendingAnimationPicID()
				if err != nil {
					return nil, err
				}
				return map[string]interface{}{"pic_id": id}, nil
			})
		},
	})

	s.registerTool(mcpTool{
		Name:        "reset_sending_animation_pic_id",
		Description: "Reset HTTP animation pic id counter.",
		InputSchema: schemaObject(nil, nil),
		Handler: func(_ map[string]interface{}) (interface{}, error) {
			return s.withClient(func(c *divoom.Client) (interface{}, error) {
				if err := c.ResetSendingAnimationPicID(); err != nil {
					return nil, err
				}
				return map[string]interface{}{"ok": true}, nil
			})
		},
	})

	s.registerTool(mcpTool{
		Name:        "send_animation",
		Description: "Send animation frames with base64-encoded RGB24 frame bytes.",
		InputSchema: schemaObject(
			map[string]interface{}{
				"width":            schemaInteger("Frame width (16, 32, 64)"),
				"id":               schemaInteger("Animation pic id"),
				"speed_msecs":      schemaIntegerArray("Frame durations in ms"),
				"pic_datas_base64": schemaStringArray("RGB24 frames encoded as base64"),
			},
			[]string{"width", "id", "speed_msecs", "pic_datas_base64"},
		),
		Handler: s.handleSendAnimation,
	})

	s.registerTool(mcpTool{
		Name:        "send_animation_gif_base64",
		Description: "Send animation from base64-encoded GIF bytes.",
		InputSchema: schemaObject(
			map[string]interface{}{
				"id":         schemaInteger("Animation pic id"),
				"gif_base64": schemaString("Base64 encoded GIF bytes"),
			},
			[]string{"id", "gif_base64"},
		),
		Handler: s.handleSendAnimationGIFBase64,
	})

	s.registerTool(mcpTool{
		Name:        "send_text",
		Description: "Send scrolling or static text area.",
		InputSchema: schemaObject(
			map[string]interface{}{
				"id":    schemaInteger("Text id"),
				"x":     schemaInteger("X position"),
				"y":     schemaInteger("Y position"),
				"dir":   schemaInteger("0=left, 1=right"),
				"font":  schemaInteger("Font index"),
				"width": schemaInteger("Text width"),
				"text":  schemaString("Text content"),
				"speed": schemaInteger("Scroll speed"),
				"color": schemaString("Hex color string"),
				"align": schemaInteger("1=left, 2=middle, 3=right"),
			},
			[]string{"id", "x", "y", "dir", "font", "width", "text", "speed", "color", "align"},
		),
		Handler: s.handleSendText,
	})

	s.registerTool(mcpTool{
		Name:        "clear_all_text_area",
		Description: "Clear all text areas on device.",
		InputSchema: schemaObject(nil, nil),
		Handler: func(_ map[string]interface{}) (interface{}, error) {
			return s.withClient(func(c *divoom.Client) (interface{}, error) {
				if err := c.ClearAllTextArea(); err != nil {
					return nil, err
				}
				return map[string]interface{}{"ok": true}, nil
			})
		},
	})

	s.registerTool(mcpTool{
		Name:        "send_display_list",
		Description: "Call SendDisplayList (currently not implemented in library).",
		InputSchema: schemaObject(nil, nil),
		Handler: func(_ map[string]interface{}) (interface{}, error) {
			return s.withClient(func(c *divoom.Client) (interface{}, error) {
				if err := c.SendDisplayList(); err != nil {
					return nil, err
				}
				return map[string]interface{}{"ok": true}, nil
			})
		},
	})

	s.registerTool(clientIntTool("select_channel", "Select channel index.", "channel", func(c *divoom.Client, v int) error {
		return c.SelectChannel(divoom.Channel(v))
	}))
	s.registerTool(clientIntTool("custom_channel", "Select custom channel page index.", "custom_index", func(c *divoom.Client, v int) error {
		return c.CustomChannel(divoom.CustomIdx(v))
	}))
	s.registerTool(clientIntTool("visualizer_channel", "Select visualizer channel position.", "eq_position", func(c *divoom.Client, v int) error {
		return c.VisualizerChannel(v)
	}))
	s.registerTool(clientIntTool("cloud_channel", "Select cloud channel index.", "index", func(c *divoom.Client, v int) error {
		return c.CloudChannel(divoom.CloudChannelIdx(v))
	}))

	s.registerTool(mcpTool{
		Name:        "get_current_channel",
		Description: "Get current channel index.",
		InputSchema: schemaObject(nil, nil),
		Handler: func(_ map[string]interface{}) (interface{}, error) {
			return s.withClient(func(c *divoom.Client) (interface{}, error) {
				ch, err := c.GetCurrentChannel()
				if err != nil {
					return nil, err
				}
				return map[string]interface{}{"channel": int(ch)}, nil
			})
		},
	})

	s.registerTool(clientIntTool("select_faces_channel", "Select watchface channel by face clock id.", "clock_id", func(c *divoom.Client, v int) error {
		return c.SelectFacesChannel(v)
	}))

	s.registerTool(mcpTool{
		Name:        "get_select_face_id",
		Description: "Get selected face id and brightness.",
		InputSchema: schemaObject(nil, nil),
		Handler: func(_ map[string]interface{}) (interface{}, error) {
			return s.withClient(func(c *divoom.Client) (interface{}, error) {
				face, err := c.GetSelectFaceID()
				if err != nil {
					return nil, err
				}
				return map[string]interface{}{"face": face}, nil
			})
		},
	})

	s.registerTool(clientIntTool("set_brightness", "Set device brightness (0-100).", "brightness", func(c *divoom.Client, v int) error {
		return c.SetBrightness(v)
	}))

	s.registerTool(mcpTool{
		Name:        "get_all_setting",
		Description: "Get all device settings map.",
		InputSchema: schemaObject(nil, nil),
		Handler: func(_ map[string]interface{}) (interface{}, error) {
			return s.withClient(func(c *divoom.Client) (interface{}, error) {
				ret, err := c.GetAllSetting()
				if err != nil {
					return nil, err
				}
				return map[string]interface{}{"settings": ret}, nil
			})
		},
	})

	s.registerTool(mcpTool{
		Name:        "weather_area_setting",
		Description: "Set weather area longitude/latitude.",
		InputSchema: schemaObject(
			map[string]interface{}{
				"longitude": schemaString("Longitude"),
				"latitude":  schemaString("Latitude"),
			},
			[]string{"longitude", "latitude"},
		),
		Handler: func(args map[string]interface{}) (interface{}, error) {
			longitude, err := getString(args, "longitude", true)
			if err != nil {
				return nil, err
			}
			latitude, err := getString(args, "latitude", true)
			if err != nil {
				return nil, err
			}
			return s.withClient(func(c *divoom.Client) (interface{}, error) {
				if err := c.WeatherAreaSetting(longitude, latitude); err != nil {
					return nil, err
				}
				return map[string]interface{}{"ok": true}, nil
			})
		},
	})

	s.registerTool(mcpTool{
		Name:        "set_time_zone",
		Description: "Set timezone value string.",
		InputSchema: schemaObject(
			map[string]interface{}{
				"timezone": schemaString("Timezone value"),
			},
			[]string{"timezone"},
		),
		Handler: func(args map[string]interface{}) (interface{}, error) {
			timezone, err := getString(args, "timezone", true)
			if err != nil {
				return nil, err
			}
			return s.withClient(func(c *divoom.Client) (interface{}, error) {
				if err := c.SetTimeZone(timezone); err != nil {
					return nil, err
				}
				return map[string]interface{}{"ok": true}, nil
			})
		},
	})

	s.registerTool(mcpTool{
		Name:        "set_system_time",
		Description: "Set device UTC time string.",
		InputSchema: schemaObject(
			map[string]interface{}{
				"utc_time": schemaString("UTC time string"),
			},
			[]string{"utc_time"},
		),
		Handler: func(args map[string]interface{}) (interface{}, error) {
			utcTime, err := getString(args, "utc_time", true)
			if err != nil {
				return nil, err
			}
			return s.withClient(func(c *divoom.Client) (interface{}, error) {
				if err := c.SystemTime(utcTime); err != nil {
					return nil, err
				}
				return map[string]interface{}{"ok": true}, nil
			})
		},
	})

	s.registerTool(clientBoolTool("screen_switch", "Turn screen on/off.", "on", func(c *divoom.Client, v bool) error {
		return c.ScreenSwitch(v)
	}))

	s.registerTool(mcpTool{
		Name:        "get_device_time",
		Description: "Get current device time.",
		InputSchema: schemaObject(nil, nil),
		Handler: func(_ map[string]interface{}) (interface{}, error) {
			return s.withClient(func(c *divoom.Client) (interface{}, error) {
				ret, err := c.GetDeviceTime()
				if err != nil {
					return nil, err
				}
				return map[string]interface{}{"device_time": ret}, nil
			})
		},
	})

	s.registerTool(clientIntTool("set_temperature_mode", "Set temperature mode (0=celsius, 1=fahrenheit).", "mode", func(c *divoom.Client, v int) error {
		return c.SetTemperatureMode(divoom.TempMode(v))
	}))
	s.registerTool(clientIntTool("set_rotation_angle", "Set rotation angle (0,1,2,3).", "mode", func(c *divoom.Client, v int) error {
		return c.SetRotationAngle(divoom.RotationAngle(v))
	}))
	s.registerTool(clientIntTool("set_mirror_mode", "Set mirror mode (0=disable,1=enable).", "mode", func(c *divoom.Client, v int) error {
		return c.SetMirrorMode(divoom.MirrorMode(v))
	}))
	s.registerTool(clientIntTool("set_hour_mode", "Set hour mode (0=12h,1=24h).", "mode", func(c *divoom.Client, v int) error {
		return c.SetHourMode(divoom.HourMode(v))
	}))
	s.registerTool(clientBoolTool("set_high_light_mode", "Set high light mode on/off.", "on", func(c *divoom.Client, v bool) error {
		return c.SetHighLightMode(v)
	}))

	s.registerTool(mcpTool{
		Name:        "set_white_balance",
		Description: "Set RGB white balance (0-100 each).",
		InputSchema: schemaObject(
			map[string]interface{}{
				"r": schemaInteger("Red value"),
				"g": schemaInteger("Green value"),
				"b": schemaInteger("Blue value"),
			},
			[]string{"r", "g", "b"},
		),
		Handler: func(args map[string]interface{}) (interface{}, error) {
			r, err := getInt(args, "r", true)
			if err != nil {
				return nil, err
			}
			g, err := getInt(args, "g", true)
			if err != nil {
				return nil, err
			}
			b, err := getInt(args, "b", true)
			if err != nil {
				return nil, err
			}
			return s.withClient(func(c *divoom.Client) (interface{}, error) {
				if err := c.SetWhiteBalance(r, g, b); err != nil {
					return nil, err
				}
				return map[string]interface{}{"ok": true}, nil
			})
		},
	})

	s.registerTool(mcpTool{
		Name:        "set_countdown_tool",
		Description: "Set countdown timer by seconds and start status.",
		InputSchema: schemaObject(
			map[string]interface{}{
				"seconds": schemaInteger("Countdown duration in seconds"),
				"start":   schemaBoolean("Start flag"),
			},
			[]string{"seconds", "start"},
		),
		Handler: func(args map[string]interface{}) (interface{}, error) {
			seconds, err := getInt(args, "seconds", true)
			if err != nil {
				return nil, err
			}
			start, err := getBool(args, "start", true)
			if err != nil {
				return nil, err
			}
			return s.withClient(func(c *divoom.Client) (interface{}, error) {
				if err := c.SetCountdownTool(time.Duration(seconds)*time.Second, start); err != nil {
					return nil, err
				}
				return map[string]interface{}{"ok": true}, nil
			})
		},
	})

	s.registerTool(clientIntTool("set_stopwatch_tool", "Set stopwatch status (0=stop,1=start,2=reset).", "status", func(c *divoom.Client, v int) error {
		return c.SetStopwatchTool(divoom.StopwatchStatus(v))
	}))

	s.registerTool(mcpTool{
		Name:        "set_scoreboard_tool",
		Description: "Set scoreboard scores.",
		InputSchema: schemaObject(
			map[string]interface{}{
				"red":  schemaInteger("Red score"),
				"blue": schemaInteger("Blue score"),
			},
			[]string{"red", "blue"},
		),
		Handler: func(args map[string]interface{}) (interface{}, error) {
			red, err := getInt(args, "red", true)
			if err != nil {
				return nil, err
			}
			blue, err := getInt(args, "blue", true)
			if err != nil {
				return nil, err
			}
			return s.withClient(func(c *divoom.Client) (interface{}, error) {
				if err := c.SetScoreboardTool(red, blue); err != nil {
					return nil, err
				}
				return map[string]interface{}{"ok": true}, nil
			})
		},
	})

	s.registerTool(clientBoolTool("set_noise_tool", "Set noise status on/off.", "on", func(c *divoom.Client, v bool) error {
		return c.SetNoiseTool(v)
	}))
}

func clientIntTool(name, desc, key string, fn func(*divoom.Client, int) error) mcpTool {
	return mcpTool{
		Name:        name,
		Description: desc,
		InputSchema: schemaObject(map[string]interface{}{
			key: schemaInteger("Integer value"),
		}, []string{key}),
		Handler: func(args map[string]interface{}) (interface{}, error) {
			return nil, fmt.Errorf("%s is not bound", name)
		},
	}
}

func clientBoolTool(name, desc, key string, fn func(*divoom.Client, bool) error) mcpTool {
	return mcpTool{
		Name:        name,
		Description: desc,
		InputSchema: schemaObject(map[string]interface{}{
			key: schemaBoolean("Boolean value"),
		}, []string{key}),
		Handler: func(args map[string]interface{}) (interface{}, error) {
			return nil, fmt.Errorf("%s is not bound", name)
		},
	}
}

func (s *mcpServer) bindGeneratedTools() {
	for name, tool := range s.tools {
		switch name {
		case "select_channel":
			s.tools[name] = bindIntTool(tool, s, "channel", func(c *divoom.Client, v int) error { return c.SelectChannel(divoom.Channel(v)) })
		case "custom_channel":
			s.tools[name] = bindIntTool(tool, s, "custom_index", func(c *divoom.Client, v int) error { return c.CustomChannel(divoom.CustomIdx(v)) })
		case "visualizer_channel":
			s.tools[name] = bindIntTool(tool, s, "eq_position", func(c *divoom.Client, v int) error { return c.VisualizerChannel(v) })
		case "cloud_channel":
			s.tools[name] = bindIntTool(tool, s, "index", func(c *divoom.Client, v int) error { return c.CloudChannel(divoom.CloudChannelIdx(v)) })
		case "select_faces_channel":
			s.tools[name] = bindIntTool(tool, s, "clock_id", func(c *divoom.Client, v int) error { return c.SelectFacesChannel(v) })
		case "set_brightness":
			s.tools[name] = bindIntTool(tool, s, "brightness", func(c *divoom.Client, v int) error { return c.SetBrightness(v) })
		case "screen_switch":
			s.tools[name] = bindBoolTool(tool, s, "on", func(c *divoom.Client, v bool) error { return c.ScreenSwitch(v) })
		case "set_temperature_mode":
			s.tools[name] = bindIntTool(tool, s, "mode", func(c *divoom.Client, v int) error { return c.SetTemperatureMode(divoom.TempMode(v)) })
		case "set_rotation_angle":
			s.tools[name] = bindIntTool(tool, s, "mode", func(c *divoom.Client, v int) error { return c.SetRotationAngle(divoom.RotationAngle(v)) })
		case "set_mirror_mode":
			s.tools[name] = bindIntTool(tool, s, "mode", func(c *divoom.Client, v int) error { return c.SetMirrorMode(divoom.MirrorMode(v)) })
		case "set_hour_mode":
			s.tools[name] = bindIntTool(tool, s, "mode", func(c *divoom.Client, v int) error { return c.SetHourMode(divoom.HourMode(v)) })
		case "set_high_light_mode":
			s.tools[name] = bindBoolTool(tool, s, "on", func(c *divoom.Client, v bool) error { return c.SetHighLightMode(v) })
		case "set_stopwatch_tool":
			s.tools[name] = bindIntTool(tool, s, "status", func(c *divoom.Client, v int) error { return c.SetStopwatchTool(divoom.StopwatchStatus(v)) })
		case "set_noise_tool":
			s.tools[name] = bindBoolTool(tool, s, "on", func(c *divoom.Client, v bool) error { return c.SetNoiseTool(v) })
		}
	}
}

func bindIntTool(tool mcpTool, s *mcpServer, key string, fn func(*divoom.Client, int) error) mcpTool {
	tool.Handler = func(args map[string]interface{}) (interface{}, error) {
		v, err := getInt(args, key, true)
		if err != nil {
			return nil, err
		}
		return s.withClient(func(c *divoom.Client) (interface{}, error) {
			if err := fn(c, v); err != nil {
				return nil, err
			}
			return map[string]interface{}{"ok": true}, nil
		})
	}
	return tool
}

func bindBoolTool(tool mcpTool, s *mcpServer, key string, fn func(*divoom.Client, bool) error) mcpTool {
	tool.Handler = func(args map[string]interface{}) (interface{}, error) {
		v, err := getBool(args, key, true)
		if err != nil {
			return nil, err
		}
		return s.withClient(func(c *divoom.Client) (interface{}, error) {
			if err := fn(c, v); err != nil {
				return nil, err
			}
			return map[string]interface{}{"ok": true}, nil
		})
	}
	return tool
}

func (s *mcpServer) handleConnectDevice(args map[string]interface{}) (interface{}, error) {
	devs, err := divoom.FindDevice()
	if err != nil {
		return nil, err
	}
	if len(devs) == 0 {
		return nil, fmt.Errorf("no device found")
	}

	var target *divoom.Device
	if ip, err := getString(args, "ip", false); err == nil && ip != "" {
		for _, d := range devs {
			if d.DevicePrivateIP == ip {
				target = d
				break
			}
		}
		if target == nil {
			return nil, fmt.Errorf("device with ip %q not found", ip)
		}
	} else if deviceID, err := getInt(args, "device_id", false); err == nil {
		for _, d := range devs {
			if d.DeviceID == deviceID {
				target = d
				break
			}
		}
		if target == nil {
			return nil, fmt.Errorf("device id %d not found", deviceID)
		}
	} else if idx, err := getInt(args, "index", false); err == nil {
		if idx < 0 || idx >= len(devs) {
			return nil, fmt.Errorf("index out of range: %d", idx)
		}
		target = devs[idx]
	} else {
		target = devs[0]
	}

	s.mu.Lock()
	s.device = target
	s.client = divoom.NewClient(target)
	s.mu.Unlock()
	return map[string]interface{}{"device": target}, nil
}

func (s *mcpServer) handleSendAnimation(args map[string]interface{}) (interface{}, error) {
	width, err := getInt(args, "width", true)
	if err != nil {
		return nil, err
	}
	id, err := getInt(args, "id", true)
	if err != nil {
		return nil, err
	}
	speeds, err := getIntSlice(args, "speed_msecs", true)
	if err != nil {
		return nil, err
	}
	encodedFrames, err := getStringSlice(args, "pic_datas_base64", true)
	if err != nil {
		return nil, err
	}

	frameBytes := make([][]byte, len(encodedFrames))
	for i, item := range encodedFrames {
		raw, decodeErr := base64.StdEncoding.DecodeString(item)
		if decodeErr != nil {
			return nil, fmt.Errorf("invalid base64 at pic_datas_base64[%d]: %w", i, decodeErr)
		}
		frameBytes[i] = raw
	}

	return s.withClient(func(c *divoom.Client) (interface{}, error) {
		if err := c.SendAnimation(width, id, speeds, frameBytes); err != nil {
			return nil, err
		}
		return map[string]interface{}{"ok": true}, nil
	})
}

func (s *mcpServer) handleSendAnimationGIFBase64(args map[string]interface{}) (interface{}, error) {
	id, err := getInt(args, "id", true)
	if err != nil {
		return nil, err
	}
	encodedGIF, err := getString(args, "gif_base64", true)
	if err != nil {
		return nil, err
	}
	rawGIF, err := base64.StdEncoding.DecodeString(encodedGIF)
	if err != nil {
		return nil, fmt.Errorf("invalid gif_base64: %w", err)
	}
	gifImg, err := gif.DecodeAll(bytes.NewReader(rawGIF))
	if err != nil {
		return nil, fmt.Errorf("invalid gif data: %w", err)
	}

	return s.withClient(func(c *divoom.Client) (interface{}, error) {
		if err := c.SendAnimationGif(id, gifImg); err != nil {
			return nil, err
		}
		return map[string]interface{}{"ok": true}, nil
	})
}

func (s *mcpServer) handleSendText(args map[string]interface{}) (interface{}, error) {
	id, err := getInt(args, "id", true)
	if err != nil {
		return nil, err
	}
	x, err := getInt(args, "x", true)
	if err != nil {
		return nil, err
	}
	y, err := getInt(args, "y", true)
	if err != nil {
		return nil, err
	}
	dir, err := getInt(args, "dir", true)
	if err != nil {
		return nil, err
	}
	font, err := getInt(args, "font", true)
	if err != nil {
		return nil, err
	}
	width, err := getInt(args, "width", true)
	if err != nil {
		return nil, err
	}
	text, err := getString(args, "text", true)
	if err != nil {
		return nil, err
	}
	speed, err := getInt(args, "speed", true)
	if err != nil {
		return nil, err
	}
	color, err := getString(args, "color", true)
	if err != nil {
		return nil, err
	}
	align, err := getInt(args, "align", true)
	if err != nil {
		return nil, err
	}

	return s.withClient(func(c *divoom.Client) (interface{}, error) {
		if err := c.SendText(id, x, y, divoom.TextDir(dir), divoom.TextFont(font), width, text, speed, color, divoom.TextAlign(align)); err != nil {
			return nil, err
		}
		return map[string]interface{}{"ok": true}, nil
	})
}

func (s *mcpServer) withClient(fn func(*divoom.Client) (interface{}, error)) (interface{}, error) {
	s.mu.RLock()
	client := s.client
	s.mu.RUnlock()
	if client == nil {
		return nil, fmt.Errorf("no connected device; call connect_device first")
	}
	return fn(client)
}

func (s *mcpServer) handleRequest(req rpcRequest) *rpcResponse {
	if req.JSONRPC != "2.0" {
		return s.newErrorResponse(req.ID, -32600, "invalid jsonrpc version", nil)
	}

	switch req.Method {
	case "initialize":
		return &rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"protocolVersion": mcpProtocolVersion,
				"capabilities": map[string]interface{}{
					"tools": map[string]interface{}{},
				},
				"serverInfo": map[string]interface{}{
					"name":    serverName,
					"version": serverVersion,
				},
			},
		}
	case "notifications/initialized":
		return nil
	case "ping":
		return &rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  map[string]interface{}{},
		}
	case "tools/list":
		tools := make([]map[string]interface{}, 0, len(s.tools))
		for _, t := range s.tools {
			tools = append(tools, map[string]interface{}{
				"name":        t.Name,
				"description": t.Description,
				"inputSchema": t.InputSchema,
			})
		}
		return &rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"tools": tools,
			},
		}
	case "tools/call":
		var p toolCallParams
		if err := json.Unmarshal(req.Params, &p); err != nil {
			return s.newErrorResponse(req.ID, -32602, "invalid tools/call params", err.Error())
		}
		tool, ok := s.tools[p.Name]
		if !ok {
			return s.newErrorResponse(req.ID, -32602, "unknown tool", p.Name)
		}
		if p.Arguments == nil {
			p.Arguments = map[string]interface{}{}
		}
		out, err := tool.Handler(p.Arguments)
		if err != nil {
			return &rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]interface{}{
					"content": []map[string]interface{}{
						{
							"type": "text",
							"text": err.Error(),
						},
					},
					"isError": true,
				},
			}
		}
		encoded, _ := json.Marshal(out)
		return &rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]interface{}{
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": string(encoded),
					},
				},
				"structuredContent": out,
			},
		}
	default:
		return s.newErrorResponse(req.ID, -32601, "method not found", req.Method)
	}
}

func (s *mcpServer) newErrorResponse(id json.RawMessage, code int, message string, data interface{}) *rpcResponse {
	return &rpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &rpcError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}

func serveStdio(s *mcpServer) error {
	s.bindGeneratedTools()

	dec := json.NewDecoder(os.Stdin)
	enc := json.NewEncoder(os.Stdout)

	for {
		var req rpcRequest
		if err := dec.Decode(&req); err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("decode error: %w", err)
		}

		resp := s.handleRequest(req)
		if resp == nil {
			continue
		}

		if err := enc.Encode(resp); err != nil {
			return fmt.Errorf("encode error: %w", err)
		}
	}
}


type sseSession struct {
	id      string
	msgChan chan []byte
}

func serveSSE(s *mcpServer, addr string) error {
	s.bindGeneratedTools()

	var mu sync.RWMutex
	sessions := make(map[string]*sseSession)

	mux := http.NewServeMux()
	mux.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		sessionID := strconv.FormatInt(time.Now().UnixNano()+int64(rand.Intn(1000)), 36)
		session := &sseSession{
			id:      sessionID,
			msgChan: make(chan []byte, 32),
		}

		mu.Lock()
		sessions[sessionID] = session
		mu.Unlock()
		defer func() {
			mu.Lock()
			delete(sessions, sessionID)
			mu.Unlock()
			close(session.msgChan)
		}()

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		fmt.Fprintf(w, "event: endpoint\ndata: /message?sessionId=%s\n\n", sessionID)
		flusher.Flush()

		ticker := time.NewTicker(25 * time.Second)
		defer ticker.Stop()

		ctx := r.Context()
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-session.msgChan:
				fmt.Fprintf(w, "data: %s\n\n", msg)
				flusher.Flush()
			case <-ticker.C:
				fmt.Fprint(w, ": keep-alive\n\n")
				flusher.Flush()
			}
		}
	})

	mux.HandleFunc("/message", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		sessionID := r.URL.Query().Get("sessionId")
		if sessionID == "" {
			http.Error(w, "sessionId required", http.StatusBadRequest)
			return
		}

		mu.RLock()
		session, ok := sessions[sessionID]
		mu.RUnlock()
		if !ok {
			http.Error(w, "session not found", http.StatusNotFound)
			return
		}

		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var req rpcRequest
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		resp := s.handleRequest(req)
		if resp != nil {
			data, err := json.Marshal(resp)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			select {
			case session.msgChan <- data:
			case <-time.After(3 * time.Second):
				http.Error(w, "session queue full", http.StatusTooManyRequests)
				return
			}
		}

		w.WriteHeader(http.StatusAccepted)
	})

	log.Printf("MCP SSE server listening on http://%s (stream: /sse)", addr)
	return http.ListenAndServe(addr, mux)
}

func schemaObject(properties map[string]interface{}, required []string) map[string]interface{} {
	if properties == nil {
		properties = map[string]interface{}{}
	}
	s := map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		s["required"] = required
	}
	return s
}

func schemaInteger(description string) map[string]interface{} {
	return map[string]interface{}{"type": "integer", "description": description}
}

func schemaString(description string) map[string]interface{} {
	return map[string]interface{}{"type": "string", "description": description}
}

func schemaBoolean(description string) map[string]interface{} {
	return map[string]interface{}{"type": "boolean", "description": description}
}

func schemaIntegerArray(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "array",
		"description": description,
		"items": map[string]interface{}{
			"type": "integer",
		},
	}
}

func schemaStringArray(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "array",
		"description": description,
		"items": map[string]interface{}{
			"type": "string",
		},
	}
}

func getString(args map[string]interface{}, key string, required bool) (string, error) {
	raw, ok := args[key]
	if !ok {
		if required {
			return "", fmt.Errorf("missing required argument %q", key)
		}
		return "", fmt.Errorf("argument %q not found", key)
	}
	v, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("argument %q must be string", key)
	}
	return v, nil
}

func getBool(args map[string]interface{}, key string, required bool) (bool, error) {
	raw, ok := args[key]
	if !ok {
		if required {
			return false, fmt.Errorf("missing required argument %q", key)
		}
		return false, fmt.Errorf("argument %q not found", key)
	}
	v, ok := raw.(bool)
	if !ok {
		return false, fmt.Errorf("argument %q must be boolean", key)
	}
	return v, nil
}

func getInt(args map[string]interface{}, key string, required bool) (int, error) {
	raw, ok := args[key]
	if !ok {
		if required {
			return 0, fmt.Errorf("missing required argument %q", key)
		}
		return 0, fmt.Errorf("argument %q not found", key)
	}
	switch v := raw.(type) {
	case float64:
		return int(v), nil
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case json.Number:
		n, err := v.Int64()
		if err != nil {
			return 0, fmt.Errorf("invalid number for %q: %w", key, err)
		}
		return int(n), nil
	default:
		return 0, fmt.Errorf("argument %q must be integer", key)
	}
}

func getIntSlice(args map[string]interface{}, key string, required bool) ([]int, error) {
	raw, ok := args[key]
	if !ok {
		if required {
			return nil, fmt.Errorf("missing required argument %q", key)
		}
		return nil, fmt.Errorf("argument %q not found", key)
	}
	items, ok := raw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("argument %q must be array", key)
	}
	out := make([]int, len(items))
	for i := range items {
		num, err := getInt(map[string]interface{}{"v": items[i]}, "v", true)
		if err != nil {
			return nil, fmt.Errorf("argument %q[%d]: %w", key, i, err)
		}
		out[i] = num
	}
	return out, nil
}

func getStringSlice(args map[string]interface{}, key string, required bool) ([]string, error) {
	raw, ok := args[key]
	if !ok {
		if required {
			return nil, fmt.Errorf("missing required argument %q", key)
		}
		return nil, fmt.Errorf("argument %q not found", key)
	}
	items, ok := raw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("argument %q must be array", key)
	}
	out := make([]string, len(items))
	for i := range items {
		str, ok := items[i].(string)
		if !ok {
			return nil, fmt.Errorf("argument %q[%d] must be string", key, i)
		}
		out[i] = str
	}
	return out, nil
}
