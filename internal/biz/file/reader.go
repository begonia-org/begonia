package file

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type FileReader interface {
	Reader() (io.ReadCloser, error)
	Size() int64
	ModifyTime() int64
	Close() error
	ReadAt(p []byte, offset int64) (n int, err error)
	Name() string
}
type FileVersionReader interface {
	FileReader
	Version() string
	Author() string
}

type FileReaderImpl struct {
	*os.File
	info fs.FileInfo
}

func (f *FileReaderImpl) Reader() (io.ReadCloser, error) {
	return f.File, nil
}
func (f *FileReaderImpl) Size() int64 {
	return f.info.Size()
}
func (f *FileReaderImpl) ModifyTime() int64 {
	// info, err := f.File.Stat()

	return f.info.ModTime().Unix()
}
func (f *FileReaderImpl) Close() error {
	return f.File.Close()
}
func (f *FileReaderImpl) ReadAt(p []byte, offset int64) (n int, err error) {
	return f.File.ReadAt(p, offset)

}
func (f *FileReaderImpl) Name() string {
	return f.File.Name()
}
func NewFileReader(path string) (FileReader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		return nil, err

	}
	return &FileReaderImpl{File: file, info: info}, nil
}

type ReaderAtCloser struct {
	io.ReadCloser
}
type fileVersionReaderImpl struct {
	*git.Repository
	version string
	name    string
	file    *object.File
	commit  *object.Commit
	path    string
}

func (f *fileVersionReaderImpl) Reader() (io.ReadCloser, error) {
	// 通过commit ID找到commit对象
	commitHash := plumbing.NewHash(f.version)
	commit, err := f.Repository.CommitObject(commitHash)
	if err != nil {
		return nil, err
	}
	// 从commit获取树
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	// 在树中查找文件
	// ent,_:=tree.FindEntry(fileName)
	file, err := tree.File(f.name)
	if err != nil {
		return nil, err
	}
	return file.Reader()
}

// func (f *fileVersionReaderImpl) Commit() (*object.Commit, error) {
// 	commitHash := plumbing.NewHash(f.version)
// 	return f.Repository.CommitObject(commitHash)
// }

func (f *fileVersionReaderImpl) Size() int64 {
	return f.file.Size
}
func (f *fileVersionReaderImpl) ModifyTime() int64 {
	return f.commit.Committer.When.Unix()
}
func (f *fileVersionReaderImpl) Version() string {
	return f.version
}
func (f *fileVersionReaderImpl) Author() string {
	return f.commit.Author.Name
}
func (f *fileVersionReaderImpl) Close() error {
	return nil
}
func (f *fileVersionReaderImpl) Name() string {
	return f.path

}
func (f *fileVersionReaderImpl) ReadAt(p []byte, offset int64) (n int, err error) {
	reader, err := f.Reader()
	if err != nil {
		return 0, err
	}
	defer reader.Close()
	// 读取并丢弃offset之前的数据
	if _, err := io.CopyN(io.Discard, reader, offset); err != nil && err != io.EOF {
		return 0, fmt.Errorf("failed to discard offset bytes: %w", err)
	}

	// 现在从offset开始读取length长度的数据
	buf := make([]byte, len(p))
	// var size int
	if _, err = io.ReadFull(reader, buf); err != nil && err != io.EOF {
		return 0, fmt.Errorf("failed to read data: %w", err)
	}
	return copy(p, buf), nil

}
func NewFileVersionReader(path string, version string) (FileVersionReader, error) {
	dir := filepath.Dir(path)
	fileName := filepath.Base(path)
	repo, err := git.PlainOpen(dir)
	if err != nil {
		log.Printf("failed to open git repository,%s: %v",dir, err)
		return nil, err

	}
	if version == "" || version == "latest" {
		ref, err := repo.Head()
		if err != nil {
			return nil, err
		}
		version = ref.Hash().String()

	}
	commitHash := plumbing.NewHash(version)
	commit, err := repo.CommitObject(commitHash)
	if err != nil {
		return nil, err

	}
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}
	file, err := tree.File(fileName)
	if err != nil {
		return nil, err

	}
	return &fileVersionReaderImpl{Repository: repo, version: version, name: fileName, file: file, commit: commit, path: path}, nil
	// return &fileVersionReaderImpl{Repository: repo, version: version, name: name}, nil
}
