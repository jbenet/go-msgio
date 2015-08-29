package msgio

import (
	"io"
)

// LimitedReader wraps an io.Reader with a msgio framed reader. The LimitedReader
// will return a reader which will io.EOF when the msg length is done.
func LimitedReader(r io.Reader) io.Reader {
	l := int64(0)
	lbuf := make([]byte, lengthSize)
	if _, err := io.ReadFull(r, lbuf); err == nil {
		l = int64(NBO.Uint32(lbuf))
	}
	return io.LimitReader(r, l)
}
