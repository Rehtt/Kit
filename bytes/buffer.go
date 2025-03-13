package bytes

import "io"

type ByteBuffer struct {
	buffer []byte
	index  int
}

func MakeByteBuffer(buffer []byte) ByteBuffer {
	return ByteBuffer{
		buffer: buffer,
		index:  0,
	}
}

func (bb *ByteBuffer) Reset() {
	bb.index = 0
}

func (bb *ByteBuffer) Len() int {
	return len(bb.buffer)
}

func (bb *ByteBuffer) Position() int {
	return bb.index
}

func (bb *ByteBuffer) Bytes() []byte {
	return bb.buffer
}

func (bb *ByteBuffer) String() string {
	return ToString(bb.buffer)
}

func (bb *ByteBuffer) Read(buffer []byte) (int, error) {
	if len(buffer) == 0 {
		return 0, nil
	}

	if bb.index >= bb.Len() {
		return 0, io.EOF
	}

	last := copy(buffer, bb.buffer[bb.index:])
	bb.index += last
	return last, nil
}

func (bb *ByteBuffer) Write(buffer []byte) (int, error) {
	bb.buffer = append(bb.buffer[:bb.index], buffer...)
	return len(buffer), nil
}

func (bb *ByteBuffer) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	default:
	case io.SeekStart:
		bb.index = int(offset)
	case io.SeekCurrent:
		bb.index += int(offset)
	case io.SeekEnd:
		bb.index = bb.Len() - 1 - int(offset)
	}
	return int64(bb.index), nil
}

func (bb *ByteBuffer) Close() error {
	// 不清空底层分配
	bb.buffer = bb.buffer[:0]
	bb.index = 0
	return nil
}
