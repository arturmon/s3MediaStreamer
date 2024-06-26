package logs

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"crypto/rand"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"sync"
)

type UDPWriter struct {
	GelfWriter
	CompressionLevel int // one of the consts from compress/flate
	CompressionType  CompressType
}

// What compression type the writer should use when sending messages
// to the graylog2 server.
type CompressType int

const (
	CompressGzip CompressType = iota
	CompressZlib
	CompressNone
)

// Used to control GELF chunking.  Should be less than (MTU - len(UDP
// header)).
//
// TODO: generate dynamically using Path MTU Discovery?
const (
	ChunkSize         = 1420
	chunkedHeaderLen  = 12
	chunkedDataLen    = ChunkSize - chunkedHeaderLen
	defaultBufferSize = 1024
	maxChunks         = 128
)
const (
	magicChunkedByte1 = 0x1e
	magicChunkedByte2 = 0x0f
	magicZlibByte     = 0x78
	magicGzipByte1    = 0x1f
	magicGzipByte2    = 0x8b
)

var (
	magicChunked = []byte{magicChunkedByte1, magicChunkedByte2}
	magicZlib    = []byte{magicZlibByte}
	magicGzip    = []byte{magicGzipByte1, magicGzipByte2}
)

// numChunks returns the number of GELF chunks necessary to transmit
// the given compressed buffer.
func numChunks(b []byte) int {
	lenB := len(b)
	if lenB <= ChunkSize {
		return 1
	}
	return len(b)/chunkedDataLen + 1
}

// New returns a new GELF Writer.  This writer can be used to send the
// output of the standard Go log functions to a central GELF server by
// passing it to log.SetOutput().
func NewUDPWriter(addr string, appName string) (*UDPWriter, error) {
	var err error
	w := new(UDPWriter)
	w.CompressionLevel = flate.BestSpeed

	if w.conn, err = net.Dial("udp", addr); err != nil {
		return nil, err
	}
	if w.hostname, err = os.Hostname(); err != nil {
		return nil, err
	}
	w.Facility = path.Base(os.Args[0])
	w.hostname = appName + "-" + w.hostname

	return w, nil
}

// writes the gzip compressed byte array to the connection as a series
// of GELF chunked messages.  The format is documented at
// http://docs.graylog.org/en/2.1/pages/gelf.html as:
//
//	2-byte magic (0x1e 0x0f), 8 byte id, 1 byte sequence id, 1 byte
//	total, chunk-data
func (w *GelfWriter) writeChunked(zBytes []byte) error {
	const (
		ChunkSize        = 1420
		chunkedHeaderLen = 12
		chunkedDataLen   = ChunkSize - chunkedHeaderLen
		maxChunks        = 128
	)

	magicChunked := []byte{0x1e, 0x0f} // define magicChunked here

	b := make([]byte, 0, ChunkSize)
	buf := bytes.NewBuffer(b)
	nChunksI := numChunks(zBytes)
	if nChunksI > maxChunks {
		return fmt.Errorf("msg too large, would need %d chunks", nChunksI)
	}
	nChunks := uint8(nChunksI)

	msgId := make([]byte, 8)
	n, err := io.ReadFull(rand.Reader, msgId)
	if err != nil || n != 8 {
		return fmt.Errorf("rand.Reader: %d/%w", n, err)
	}

	bytesLeft := len(zBytes)
	for i := uint8(0); i < nChunks; i++ {
		buf.Reset()

		buf.Write(magicChunked) // use the local variable here.
		buf.Write(msgId)
		buf.WriteByte(i)
		buf.WriteByte(nChunks)

		chunkLen := chunkedDataLen
		if chunkLen > bytesLeft {
			chunkLen = bytesLeft
		}
		off := int(i) * chunkedDataLen
		chunk := zBytes[off : off+chunkLen]
		buf.Write(chunk)

		n, err = w.conn.Write(buf.Bytes())
		if err != nil {
			return fmt.Errorf("Write (chunk %d/%d): %w", i, nChunks, err)
		}
		if n != len(buf.Bytes()) {
			return fmt.Errorf("Write len: (chunk %d/%d) (%d/%d)", i, nChunks, n, len(buf.Bytes()))
		}

		bytesLeft -= chunkLen
	}

	if bytesLeft != 0 {
		return fmt.Errorf("error: %d bytes left after sending", bytesLeft)
	}
	return nil
}

// 1k bytes buffer by default.
var bufPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, defaultBufferSize))
	},
}

func newBuffer() *bytes.Buffer {
	b, ok := bufPool.Get().(*bytes.Buffer)
	if !ok {
		// Handle the case where the pool is empty or the type assertion fails
		b = newBuffer()
	}
	b.Reset() // Reset the buffer for reuse
	return bytes.NewBuffer(nil)
}

// WriteMessage sends the specified message to the GELF server
// specified in the call to New().  It assumes all the fields are
// filled out appropriately.  In general, clients will want to use
// Write, rather than WriteMessage.
func (w *UDPWriter) WriteMessage(m *Message) error {
	mBuf := newBuffer()
	defer bufPool.Put(mBuf)
	if err := m.MarshalJSONBuf(mBuf); err != nil {
		return err
	}
	mBytes := mBuf.Bytes()

	var (
		zBuf   *bytes.Buffer
		zBytes []byte
		zw     io.WriteCloser
		err    error // Declare err here to avoid shadowing
	)

	switch w.CompressionType {
	case CompressGzip:
		zBuf = newBuffer()
		defer bufPool.Put(zBuf)
		zw, err = gzip.NewWriterLevel(zBuf, w.CompressionLevel)
	case CompressZlib:
		zBuf = newBuffer()
		defer bufPool.Put(zBuf)
		zw, err = zlib.NewWriterLevel(zBuf, w.CompressionLevel)
	case CompressNone:
		zBytes = mBytes
	default:
		return fmt.Errorf("unknown compression type %d", w.CompressionType)
	}

	if err != nil {
		return err
	}

	if zw != nil {
		if _, err = zw.Write(mBytes); err != nil {
			zw.Close()
			return err
		}
		zw.Close()
		zBytes = zBuf.Bytes()
	}

	if numChunks(zBytes) > 1 {
		return w.writeChunked(zBytes)
	}

	n, err := w.conn.Write(zBytes)
	if err != nil {
		return err
	}
	if n != len(zBytes) {
		return fmt.Errorf("bad write (%d/%d)", n, len(zBytes))
	}

	return nil
}

func (w *UDPWriter) Write(p []byte) (int, error) {
	// 1 for the function that called us.
	file, line := getCallerIgnoringLogMulti(1)

	jsonData, err := decodeJSONData(p)
	if err != nil {
		// If decoding fails, construct message using the existing logic
		m := constructMessage(p, w.hostname, w.Facility, file, line)
		if err = w.WriteMessage(m); err != nil {
			return 0, err
		}
	} else {
		// If decoding succeeds, use extracted JSON fields for constructing the message
		m := constructMessageFromJSON(w.hostname, w.Facility, jsonData)

		if err = w.WriteMessage(m); err != nil {
			return 0, err
		}
	}

	return len(p), nil
}
