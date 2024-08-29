package logs

import (
	"errors"
	"net"
	"testing"
	"time"
)

// MockConn is a mock implementation of net.Conn for testing purposes.
type MockConn struct {
	closeErr error
}

// Close simulates closing the connection and returns the predefined error.
func (m *MockConn) Close() error {
	return m.closeErr
}

// Read is a placeholder for the net.Conn Read method.
func (m *MockConn) Read(b []byte) (int, error) {
	return 0, nil
}

// Write is a placeholder for the net.Conn Write method.
func (m *MockConn) Write(b []byte) (int, error) {
	return 0, nil
}

// LocalAddr is a placeholder for the net.Conn LocalAddr method.
func (m *MockConn) LocalAddr() net.Addr {
	return nil
}

// RemoteAddr is a placeholder for the net.Conn RemoteAddr method.
func (m *MockConn) RemoteAddr() net.Addr {
	return nil
}

// SetDeadline is a placeholder for the net.Conn SetDeadline method.
func (m *MockConn) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline is a placeholder for the net.Conn SetReadDeadline method.
func (m *MockConn) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline is a placeholder for the net.Conn SetWriteDeadline method.
func (m *MockConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func TestGelfWriter_Close(t *testing.T) {
	// Test case: Normal close
	mockConn := &MockConn{}
	writer := &GelfWriter{
		conn: mockConn,
	}

	err := writer.Close()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Test case: Error on close
	mockConnWithError := &MockConn{closeErr: errors.New("close error")}
	writerWithError := &GelfWriter{
		conn: mockConnWithError,
	}

	err = writerWithError.Close()
	if err == nil || err.Error() != "close error" {
		t.Fatalf("expected 'close error', got %v", err)
	}

	// Test case: No connection to close
	writerWithoutConn := &GelfWriter{}

	err = writerWithoutConn.Close()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
