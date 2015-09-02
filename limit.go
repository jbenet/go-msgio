package msgio

import (
	"io"
)

// LimitedReader wraps an io.Reader with a msgio framed reader. The LimitedReader
// will return a reader which will io.EOF when the msg length is done.
func LimitedReader(r io.Reader) (io.Reader, error) {
	l, err := ReadLen(r, nil)
	return io.LimitReader(r, int64(l)), err
}
