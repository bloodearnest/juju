package jujutest

import (
	"bytes"
	"net/http"
	"os"
	"time"
)

type VirtualFile struct {
	*bytes.Reader
	fc *FileContent
}

var _ http.File = (*VirtualFile)(nil)

func (f *VirtualFile) Close() error {
	return nil
}

func (f *VirtualFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

func (f *VirtualFile) Stat() (os.FileInfo, error) {
	return &VirtualFileInfo{f.fc}, nil
}

type FileContent struct {
	Name    string
	Content string
}

type VirtualFileSystem struct {
	contents []FileContent
}

var _ http.FileSystem = (*VirtualFileSystem)(nil)

func (vfs *VirtualFileSystem) Open(name string) (http.File, error) {
	for _, fc := range vfs.contents {
		if fc.Name == name {
			reader := bytes.NewReader([]byte(fc.Content))
			return &VirtualFile{reader, &fc}, nil
		}
	}
	return nil, &os.PathError{Op: "Open", Path: name, Err: os.ErrNotExist}
}

func NewVFS(contents []FileContent) http.FileSystem {
	return &VirtualFileSystem{contents}
}

type VirtualFileInfo struct {
	fc *FileContent
}

var _ os.FileInfo = (*VirtualFileInfo)(nil)

func (fi *VirtualFileInfo) Name() string {
	return fi.fc.Name
}

func (fi *VirtualFileInfo) Size() int64 {
	return int64(len(fi.fc.Content))
}

func (fi *VirtualFileInfo) ModTime() time.Time {
	return time.Now()
}

func (fi *VirtualFileInfo) Mode() os.FileMode {
	return 0660
}

func (fi *VirtualFileInfo) IsDir() bool {
	return false
}

func (fi *VirtualFileInfo) Sys() interface{} { return nil }
