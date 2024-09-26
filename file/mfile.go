package file

import (
	"io/fs"
	"os"
	"time"

	"github.com/Rehtt/Kit/bytes"
)

type MentoryFile struct {
	name    string
	buffer  bytes.ByteBuffer
	modTime time.Time
}

func MakeMentoryFile(name string, buffer []byte) *MentoryFile {
	mfile := &MentoryFile{name: name}
	mfile.Write(buffer)
	return mfile
}

func (mf *MentoryFile) Stat() (fs.FileInfo, error) {
	return mf, nil
}

func (mf *MentoryFile) Read(buffer []byte) (int, error) {
	return mf.buffer.Read(buffer)
}

func (mf *MentoryFile) Close() error {
	return nil
}

func (mf *MentoryFile) Write(buffer []byte) (n int, err error) {
	mf.modTime = time.Now()
	return mf.buffer.Write(buffer)
}

func (mf *MentoryFile) Seek(offset int64, whence int) (int64, error) {
	return mf.buffer.Seek(offset, whence)
}

func (mf *MentoryFile) Bytes() []byte {
	return mf.buffer.Bytes()
}

func (mf *MentoryFile) ChangeName(name string) {
	mf.name = name
}

func (mf *MentoryFile) Name() string       { return mf.name }
func (mf *MentoryFile) Size() int64        { return int64(mf.buffer.Len()) }
func (mf *MentoryFile) Mode() os.FileMode  { return 0666 }
func (mf *MentoryFile) ModTime() time.Time { return mf.modTime }
func (mf *MentoryFile) IsDir() bool        { return false }
func (mf *MentoryFile) Sys() any           { return nil }
