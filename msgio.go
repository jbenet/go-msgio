package msgio

import (
	"encoding/binary"
	"io"
)

// NBO is NetworkByteOrder
var NBO = binary.BigEndian

// Writer is the msgio Writer interface. It writes len-framed messages.
type Writer interface {

	// WriteMsg writes the msg in the passed in buffer.
	WriteMsg([]byte) error
}

// WriteCloser is a Writer + Closer interface. Like in `golang/pkg/io`
type WriteCloser interface {
	Writer
	io.Closer
}

// Reader is the msgio Reader interface. It reads len-framed messages.
type Reader interface {

	// ReadMsg reads the next message from the Reader.
	// The client must pass a buffer large enough, or io.ErrShortBuffer will be
	// returned.
	//
	// Warning: as of this writing, this error is destructive. the length will have
	// been read.
	ReadMsg([]byte) (int, error)
}

// ReadCloser combines a Reader and Closer.
type ReadCloser interface {
	Reader
	io.Closer
}

// ReadWriter combines a Reader and Writer.
type ReadWriter interface {
	Reader
	Writer
}

// ReadWriteCloser combines a Reader, a Writer, and Closer.
type ReadWriteCloser interface {
	Reader
	Writer
	io.Closer
}

// writer is the underlying type that implements the Writer interface.
type writer struct {
	W io.Writer
}

// NewWriter wraps an io.Writer with a msgio framed writer. The msgio.Writer
// will write the length prefix of every message written.
func NewWriter(w io.Writer) WriteCloser {
	return &writer{w}
}

func (s *writer) WriteMsg(msg []byte) (err error) {
	length := uint32(len(msg))
	if err := binary.Write(s.W, NBO, &length); err != nil {
		return err
	}
	_, err = s.W.Write(msg)
	return err
}

func (s *writer) Close() error {
	if c, ok := s.W.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// reader is the underlying type that implements the Reader interface.
type reader struct {
	R    io.Reader
	lbuf []byte
}

// NewReader wraps an io.Reader with a msgio framed reader. The msgio.Reader
// will read whole messages at a time (using the length). Assumes an equivalent
// writer on the other side.
func NewReader(r io.Reader) ReadCloser {
	return &reader{r, make([]byte, 4)}
}

// nextMsgLen reads the length of the next msg into s.lbuf, and returns it.
// WARNING: like ReadMsg, nextMsgLen is destructive. It reads from the internal
// reader.
func (s *reader) nextMsgLen() (int, error) {
	if _, err := io.ReadFull(s.R, s.lbuf); err != nil {
		return 0, err
	}
	length := int(NBO.Uint32(s.lbuf))
	return length, nil
}

func (s *reader) ReadMsg(msg []byte) (int, error) {
	length, err := s.nextMsgLen()
	if err != nil {
		return 0, err
	}

	if length < 0 || length > len(msg) {
		return 0, io.ErrShortBuffer
	}
	_, err = io.ReadFull(s.R, msg[:length])
	return length, err
}

func (s *reader) Close() error {
	if c, ok := s.R.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

// readWriter is the underlying type that implements a ReadWriter.
type readWriter struct {
	Reader
	Writer
}

// NewReadWriter wraps an io.ReadWriter with a msgio.ReadWriter. Writing
// and Reading will be appropriately framed.
func NewReadWriter(rw io.ReadWriter) ReadWriter {
	return &readWriter{
		Reader: NewReader(rw),
		Writer: NewWriter(rw),
	}
}

func (rw *readWriter) Close() error {
	var errs []error

	if w, ok := rw.Writer.(WriteCloser); ok {
		if err := w.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if r, ok := rw.Reader.(ReadCloser); ok {
		if err := r.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return multiErr(errs)
	}
	return nil
}

// multiErr is a util to return multiple errors
type multiErr []error

func (m multiErr) Error() string {
	if len(m) == 0 {
		return "no errors"
	}

	s := "Multiple errors: "
	for i, e := range m {
		if i != 0 {
			s += ", "
		}
		s += e.Error()
	}
	return s
}
