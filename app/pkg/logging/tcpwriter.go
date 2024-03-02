package logging

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

const (
	DefaultMaxReconnect   = 3
	DefaultReconnectDelay = 1
)

type TCPWriter struct {
	GelfWriter
	mu             sync.Mutex
	MaxReconnect   int
	ReconnectDelay time.Duration
}

func NewTCPWriter(addr string, appName string) (*TCPWriter, error) {
	var err error
	w := new(TCPWriter)
	w.MaxReconnect = DefaultMaxReconnect
	w.ReconnectDelay = DefaultReconnectDelay
	w.proto = "tcp"
	w.addr = addr

	if w.conn, err = net.Dial("tcp", addr); err != nil {
		return nil, err
	}
	if w.hostname, err = os.Hostname(); err != nil {
		return nil, err
	}
	w.hostname = appName + "-" + w.hostname

	return w, nil
}

func (w *TCPWriter) WriteMessage(m *Message) error {
	buf := newBuffer()
	defer bufPool.Put(buf)
	messageBytes, err := m.toBytes(buf)
	if err != nil {
		return err
	}

	messageBytes = append(messageBytes, 0)

	n, err := w.writeToSocketWithReconnectAttempts(messageBytes)
	if err != nil {
		return err
	}
	if n != len(messageBytes) {
		return fmt.Errorf("bad write (%d/%d)", n, len(messageBytes))
	}

	return nil
}

func (w *TCPWriter) writeToSocketWithReconnectAttempts(zBytes []byte) (int, error) {
	var errConn error
	var i int
	var n int
	var err error // Declare err outside the loop

	w.mu.Lock()
	for i = 0; i <= w.MaxReconnect; i++ {
		errConn = nil

		if w.conn != nil {
			n, err = w.conn.Write(zBytes) // Use the existing 'err' variable
		} else {
			err = fmt.Errorf("connection was nil, will attempt reconnect") // Use the existing 'err' variable
		}
		if err != nil {
			time.Sleep(w.ReconnectDelay * time.Second)
			w.conn, errConn = net.Dial("tcp", w.addr)
		} else {
			break
		}
	}
	w.mu.Unlock()

	if i > w.MaxReconnect {
		return 0, fmt.Errorf("maximum reconnection attempts was reached; giving up")
	}
	if errConn != nil {
		return 0, fmt.Errorf("Write Failed: %s\nReconnection failed: %s", err, errConn)
	}
	return n, nil
}

func (w *TCPWriter) Write(p []byte) (int, error) {
	file, line := getCallerIgnoringLogMulti(1)

	jsonData, err := decodeJSONData(p)
	if err != nil {
		// If decoding fails, construct message using the existing logic
		m := constructMessage(p, w.hostname, w.Facility, file, line)
		if err := w.WriteMessage(m); err != nil {
			return 0, err
		}
	} else {
		// If decoding succeeds, use extracted JSON fields for constructing the message
		m := constructMessageFromJSON(w.hostname, w.Facility, jsonData)

		if err := w.WriteMessage(m); err != nil {
			return 0, err
		}
	}

	return len(p), nil
}
