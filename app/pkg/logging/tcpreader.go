package logging

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

const maxMessageQueueSize = 100

type TCPReader struct {
	listener *net.TCPListener
	conn     net.Conn
	messages chan []byte
}

type connChannels struct {
	drop    chan string
	confirm chan string
}

func newTCPReader(addr string) (*TCPReader, chan string, chan string, error) {
	var err error
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("ResolveTCPAddr('%s'): %w", addr, err)
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("ListenTCP: %w", err)
	}

	r := &TCPReader{
		listener: listener,
		messages: make(chan []byte, maxMessageQueueSize), // Make a buffered channel with at most 100 messages
	}

	closeSignal := make(chan string, 1)
	doneSignal := make(chan string, 1)

	go r.listenUntilCloseSignal(closeSignal, doneSignal)

	return r, closeSignal, doneSignal, nil
}

func (r *TCPReader) accepter(connections chan net.Conn) {
	for {
		conn, err := r.listener.Accept()
		if err != nil {
			break
		}
		connections <- conn
	}
}

func (r *TCPReader) listenUntilCloseSignal(closeSignal chan string, doneSignal chan string) {
	defer func() { doneSignal <- "done" }()
	defer r.listener.Close()

	connectionsChannel := make(chan net.Conn, 1)
	go r.accepter(connectionsChannel)

	var conns []connChannels

	for {
		select {
		case conn := <-connectionsChannel:
			r.handleNewConnection(conn, &conns)
		case sig := <-closeSignal:
			if sig == "stop" || sig == "drop" {
				r.handleStopOrDrop(sig, &conns, closeSignal, doneSignal)
			}
		default:
		}
	}
}

func (r *TCPReader) handleNewConnection(conn net.Conn, conns *[]connChannels) {
	dropSignal := make(chan string, 1)
	dropConfirm := make(chan string, 1)
	channels := connChannels{drop: dropSignal, confirm: dropConfirm}
	go handleConnection(conn, r.messages, dropSignal, dropConfirm)
	*conns = append(*conns, channels)
}

func (r *TCPReader) handleStopOrDrop(sig string, conns *[]connChannels, closeSignal chan string, doneSignal chan string) {
	if len(*conns) >= 1 {
		r.handleStopOrDropWithConnections(sig, conns)
		if sig == "stop" {
			return
		}
	} else if sig == "stop" {
		closeSignal <- "stop"
	}
	if sig == "drop" {
		doneSignal <- "done"
	}
}

func (r *TCPReader) handleStopOrDropWithConnections(sig string, conns *[]connChannels) {
	for _, s := range *conns {
		if s.drop != nil {
			s.drop <- "drop"
			<-s.confirm
			*conns = append((*conns)[:0], (*conns)[1:]...)
		}
	}
}

func (r *TCPReader) addr() string {
	return r.listener.Addr().String()
}

func handleConnection(conn net.Conn, messages chan<- []byte, dropSignal chan string, dropConfirm chan string) {
	defer func() { dropConfirm <- "done" }()
	defer conn.Close()
	reader := bufio.NewReader(conn)

	var b []byte
	var err error
	drop := false
	canDrop := false

	for {
		if err = conn.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
			// handle error appropriately
			return
		}

		if b, err = reader.ReadBytes(0); err != nil {
			if drop {
				return
			}
		} else if len(b) > 0 {
			messages <- b
			canDrop = true
			if drop {
				return
			}
		} else if drop {
			return
		}
		select {
		case sig := <-dropSignal:
			if sig == "drop" {
				drop = true
				time.Sleep(1 * time.Second)
				if canDrop {
					return
				}
			}
		default:
		}
	}
}

func (r *TCPReader) readMessage() (*Message, error) {
	b := <-r.messages

	var msg Message
	if err := json.Unmarshal(b[:len(b)-1], &msg); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}

	return &msg, nil
}

func (r *TCPReader) Close() {
	r.listener.Close()
}
