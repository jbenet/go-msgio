package msgio

import (
	"bytes"
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
