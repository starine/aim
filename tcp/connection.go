package tcp

import (
	"bufio"
	"io"
	"net"

	"github.com/starine/aim"
	"github.com/starine/aim/wire/endian"
)

// Frame Frame
type Frame struct {
	OpCode  aim.OpCode
	Payload []byte
}

// SetOpCode SetOpCode
func (f *Frame) SetOpCode(code aim.OpCode) {
	f.OpCode = code
}

// GetOpCode GetOpCode
func (f *Frame) GetOpCode() aim.OpCode {
	return f.OpCode
}

// SetPayload SetPayload
func (f *Frame) SetPayload(payload []byte) {
	f.Payload = payload
}

// GetPayload GetPayload
func (f *Frame) GetPayload() []byte {
	return f.Payload
}

// TcpConn Conn
type TcpConn struct {
	net.Conn
	rd *bufio.Reader
	wr *bufio.Writer
}

// NewConn NewConn

func NewConn(conn net.Conn) aim.Conn {
	return &TcpConn{
		Conn: conn,
		rd:   bufio.NewReaderSize(conn, 4096),
		wr:   bufio.NewWriterSize(conn, 1024),
	}
}

func NewConnWithRW(conn net.Conn, rd *bufio.Reader, wr *bufio.Writer) *TcpConn {
	return &TcpConn{
		Conn: conn,
		rd:   rd,
		wr:   wr,
	}
}

// ReadFrame ReadFrame
func (c *TcpConn) ReadFrame() (aim.Frame, error) {
	opcode, err := endian.ReadUint8(c.rd)
	if err != nil {
		return nil, err
	}
	payload, err := endian.ReadBytes(c.rd)
	if err != nil {
		return nil, err
	}
	return &Frame{
		OpCode:  aim.OpCode(opcode),
		Payload: payload,
	}, nil
}

// WriteFrame WriteFrame
func (c *TcpConn) WriteFrame(code aim.OpCode, payload []byte) error {
	return WriteFrame(c.wr, code, payload)
}

// Flush Flush
func (c *TcpConn) Flush() error {
	return c.wr.Flush()
}

// WriteFrame write a frame to w
func WriteFrame(w io.Writer, code aim.OpCode, payload []byte) error {
	if err := endian.WriteUint8(w, uint8(code)); err != nil {
		return err
	}
	if err := endian.WriteBytes(w, payload); err != nil {
		return err
	}
	return nil
}
