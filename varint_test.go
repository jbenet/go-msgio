package msgio

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestVarintReadWrite(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	writer := NewVarintWriter(buf)
	reader := NewVarintReader(buf)
	SubtestReadWrite(t, writer, reader)
}

func TestVarintReadWriteMsg(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	writer := NewVarintWriter(buf)
	reader := NewVarintReader(buf)
	SubtestReadWriteMsg(t, writer, reader)
}

func TestVarintReadWriteMsgSync(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	writer := NewVarintWriter(buf)
	reader := NewVarintReader(buf)
	SubtestReadWriteMsgSync(t, writer, reader)
}

func TestVarintWrite(t *testing.T) {

	SubtestVarintWrite(t, []byte("hello world"))
	SubtestVarintWrite(t, []byte("hello world hello world hello world"))
	SubtestVarintWrite(t, make([]byte, 1<<20))
}

func SubtestVarintWrite(t *testing.T, msg []byte) {
	buf := bytes.NewBuffer(nil)
	writer := NewVarintWriter(buf)

	if err := writer.WriteMsg(msg); err != nil {
		t.Fatal(err)
	}

	sbr := simpleByteReader{R: buf}
	n, err := binary.ReadUvarint(&sbr)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("checking varint is %d", len(msg))
	if int(n) != len(msg) {
		t.Fatalf("incorrect varint: n != %d", len(msg))
	}
}
