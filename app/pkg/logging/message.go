package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Message represents the contents of the GELF message.  It is gzipped
// before sending.
type Message struct {
	Version  string                 `json:"version"`
	Host     string                 `json:"host"`
	Short    string                 `json:"short_message"`
	Full     string                 `json:"full_message,omitempty"`
	TimeUnix float64                `json:"timestamp"`
	Level    int32                  `json:"level,omitempty"`
	Facility string                 `json:"facility,omitempty"`
	Extra    map[string]interface{} `json:"-"`
	RawExtra json.RawMessage        `json:"-"`
}

// Syslog severity levels
const (
	LOG_EMERG          = 0
	LOG_ALERT          = 1
	LOG_CRIT           = 2
	LOG_ERR            = 3
	LOG_WARNING        = 4
	LOG_NOTICE         = 5
	LOG_INFO           = 6
	LOG_DEBUG          = 7
	initialMapCapacity = 16
)

func (m *Message) MarshalJSONBuf(buf *bytes.Buffer) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	// write up until the final }
	if _, err = buf.Write(b[:len(b)-1]); err != nil {
		return err
	}
	if len(m.Extra) > 0 {
		eb, extraErr := json.Marshal(m.Extra)
		if extraErr != nil {
			return err
		}
		// merge serialized message + serialized extra map
		if extraErr = buf.WriteByte(','); extraErr != nil {
			return err
		}
		// write serialized extra bytes, without enclosing quotes
		if _, extraErr = buf.Write(eb[1 : len(eb)-1]); extraErr != nil {
			return err
		}
	}

	if len(m.RawExtra) > 0 {
		if err = buf.WriteByte(','); err != nil {
			return err
		}

		// write serialized extra bytes, without enclosing quotes
		if _, err = buf.Write(m.RawExtra[1 : len(m.RawExtra)-1]); err != nil {
			return err
		}
	}

	// write final closing quotes
	return buf.WriteByte('}')
}

func (m *Message) UnmarshalJSON(data []byte) error {
	i := make(map[string]interface{}, initialMapCapacity)
	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}
	for k, v := range i {
		if k[0] == '_' {
			if m.Extra == nil {
				m.Extra = make(map[string]interface{}, 1)
			}
			m.Extra[k] = v
			continue
		}

		ok := true
		switch k {
		case "version":
			m.Version, ok = v.(string)
		case "host":
			m.Host, ok = v.(string)
		case "short_message":
			m.Short, ok = v.(string)
		case "full_message":
			m.Full, ok = v.(string)
		case "timestamp":
			m.TimeUnix, ok = v.(float64)
		case "level":
			var level float64
			level, ok = v.(float64)
			m.Level = int32(level)
		case "facility":
			m.Facility, ok = v.(string)
		}

		if !ok {
			return fmt.Errorf("invalid type for field %s", k)
		}
	}
	return nil
}

func (m *Message) toBytes(buf *bytes.Buffer) ([]byte, error) {
	if err := m.MarshalJSONBuf(buf); err != nil {
		return nil, err
	}
	messageBytes := buf.Bytes()
	return messageBytes, nil
}

func constructMessage(p []byte, hostname string, facility string, file string, line int) *Message {
	// remove trailing and leading whitespace
	p = bytes.TrimSpace(p)

	// If there are newlines in the message, use the first line
	// for the short message and set the full message to the
	// original input.  If the input has no newlines, stick the
	// whole thing in Short.
	short := p
	full := []byte("")
	if i := bytes.IndexRune(p, '\n'); i > 0 {
		short = p[:i]
		full = p
	}

	m := &Message{
		Version:  "1.1",
		Host:     hostname,
		Short:    string(short),
		Full:     string(full),
		TimeUnix: float64(time.Now().UnixNano()) / float64(time.Second),
		Level:    LOG_INFO, // info
		Facility: facility,
		Extra: map[string]interface{}{
			"_file": file,
			"_line": line,
		},
	}

	return m
}

func constructMessageFromJSON(hostname, facility string, jsonData map[string]interface{}) *Message {
	level, ok := jsonData["level"].(string)
	if !ok {
		level = "info" // default to info if level is not present
	}

	// Convert level to GELF level
	gelfLevel := map[string]int{
		"debug":  LOG_DEBUG,
		"info":   LOG_INFO,
		"notice": LOG_NOTICE,
		"warn":   LOG_WARNING,
		"error":  LOG_ERR,
		"fatal":  LOG_CRIT,
		"panic":  LOG_ALERT,
	}[strings.ToLower(level)]

	var timeUnix float64
	if jsonDataTime, timeOk := jsonData["time"].(float64); timeOk {
		timeUnix = jsonDataTime
	} else {
		// Handle the case when jsonData["time"] is not a float64
		// You may want to log a warning or handle it based on your requirements
		timeUnix = float64(time.Now().UnixNano()) / float64(time.Second)
	}

	m := &Message{
		Version:  "1.1",
		Host:     hostname,
		Short:    fmt.Sprintf("%v", jsonData["msg"]), // assuming the message key exists in the JSON
		TimeUnix: timeUnix,
		Level:    int32(gelfLevel),
		Facility: facility,
		Extra: map[string]interface{}{
			"_file": jsonData["_file"],
			"_line": jsonData["_line"],
			"_func": jsonData["_func"],
			// Add other fields from jsonData as needed
		},
	}

	return m
}
