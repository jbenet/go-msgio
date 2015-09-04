package msgio

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"sync"
	"testing"
	"time"

	randbuf "github.com/jbenet/go-randbuf"
)

func TestReadWrite(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	writer := NewWriter(buf)
	reader := NewReader(buf)
	SubtestReadWrite(t, writer, reader)
}

func TestReadWriteMsg(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	writer := NewWriter(buf)
	reader := NewReader(buf)
	SubtestReadWriteMsg(t, writer, reader)
}

func TestReadWriteMsgSync(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	writer := NewWriter(buf)
	reader := NewReader(buf)
	SubtestReadWriteMsgSync(t, writer, reader)
}

func SubtestReadWrite(t *testing.T, writer WriteCloser, reader ReadCloser) {
	msgs := [1000][]byte{}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range msgs {
		msgs[i] = randbuf.RandBuf(r, r.Intn(1000))
		n, err := writer.Write(msgs[i])
		if err != nil {
			t.Fatal(err)
		}
		if n != len(msgs[i]) {
			t.Fatal("wrong length:", n, len(msgs[i]))
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	for i := 0; ; i++ {
		msg2 := make([]byte, 1000)
		n, err := reader.Read(msg2)
		if err != nil {
			if err == io.EOF {
				if i < len(msg2) {
					t.Error("failed to read all messages", len(msgs), i)
				}
				break
			}
			t.Error("unexpected error", err)
		}

		msg1 := msgs[i]
		msg2 = msg2[:n]
		if !bytes.Equal(msg1, msg2) {
			t.Fatal("message retrieved not equal\n", msg1, "\n\n", msg2)
		}
	}

	if err := reader.Close(); err != nil {
		t.Error(err)
	}
}

func SubtestReadWriteMsg(t *testing.T, writer WriteCloser, reader ReadCloser) {
	msgs := [1000][]byte{}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range msgs {
		msgs[i] = randbuf.RandBuf(r, r.Intn(1000))
		err := writer.WriteMsg(msgs[i])
		if err != nil {
			t.Fatal(err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	for i := 0; ; i++ {
		msg2, err := reader.ReadMsg()
		if err != nil {
			if err == io.EOF {
				if i < len(msg2) {
					t.Error("failed to read all messages", len(msgs), i)
				}
				break
			}
			t.Error("unexpected error", err)
		}

		msg1 := msgs[i]
		if !bytes.Equal(msg1, msg2) {
			t.Fatal("message retrieved not equal\n", msg1, "\n\n", msg2)
		}
	}

	if err := reader.Close(); err != nil {
		t.Error(err)
	}
}

func SubtestReadWriteMsgSync(t *testing.T, writer WriteCloser, reader ReadCloser) {
	msgs := [1000][]byte{}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range msgs {
		msgs[i] = randbuf.RandBuf(r, r.Intn(1000)+4)
		NBO.PutUint32(msgs[i][:4], uint32(i))
	}

	var wg1 sync.WaitGroup
	var wg2 sync.WaitGroup

	errs := make(chan error, 10000)
	for i := range msgs {
		wg1.Add(1)
		go func(i int) {
			defer wg1.Done()

			err := writer.WriteMsg(msgs[i])
			if err != nil {
				errs <- err
			}
		}(i)
	}

	wg1.Wait()
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < len(msgs)+1; i++ {
		wg2.Add(1)
		go func(i int) {
			defer wg2.Done()

			msg2, err := reader.ReadMsg()
			if err != nil {
				if err == io.EOF {
					if i < len(msg2) {
						errs <- fmt.Errorf("failed to read all messages", len(msgs), i)
					}
					return
				}
				errs <- fmt.Errorf("unexpected error", err)
			}

			mi := NBO.Uint32(msg2[:4])
			msg1 := msgs[mi]
			if !bytes.Equal(msg1, msg2) {
				errs <- fmt.Errorf("message retrieved not equal\n", msg1, "\n\n", msg2)
			}
		}(i)
	}

	wg2.Wait()
	close(errs)

	if err := reader.Close(); err != nil {
		t.Error(err)
	}

	for e := range errs {
		t.Error(e)
	}
}

func TestReadMaxSize(t *testing.T) {
	buf := new(bytes.Buffer)
	writer := NewWriter(buf)
	writer.WriteMsg(bytes.Repeat([]byte("x"), 11))
	rd := NewReader(buf, MaxSize(10))
	_, err := rd.ReadMsg()
	if err == nil {
		t.Fatal("should get an error")
	}
	if err != ErrMsgTooLarge {
		t.Fatal("should return ErrMsgTooLarge")
	}
}

func TestReadMaxSize_defaultSize(t *testing.T) {
	buf := new(bytes.Buffer)
	writer := NewWriter(buf)
	writer.WriteMsg(bytes.Repeat([]byte("x"), defaultMaxSize+1))
	rd := NewReader(buf)
	_, err := rd.ReadMsg()
	if err == nil {
		t.Fatal("should get an error")
	}
	if err != ErrMsgTooLarge {
		t.Fatal("should return ErrMsgTooLarge")
	}
}
