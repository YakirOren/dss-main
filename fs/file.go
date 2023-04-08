package fs

import (
	"context"
	"dss-main/storage"
	"errors"
	"github.com/yakiroren/dss-common/db"
	"github.com/yakiroren/dss-common/models"
	"io"
	"io/fs"
)

type File struct {
	reader    io.ReadCloser
	metadata  *models.FileMetadata
	datastore db.DataStore
	storage   *storage.Client
}

func (f *File) Read(p []byte) (int, error) {
	if f.reader == nil {
		readCloser, err := f.storage.ReadFragments(context.Background(), f.metadata.Fragments)
		if err != nil {
			return 0, err
		}

		f.reader = readCloser
	}

	return f.reader.Read(p)
}

func (f File) Close() error {
	if f.reader != nil {
		return f.reader.Close()
	}
	return nil
}

func (f File) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (f File) Readdir(count int) ([]fs.FileInfo, error) {
	if !f.metadata.IsDir() {
		return nil, errors.New("file is not a directory")
	}

	listFiles, err := f.datastore.ListFiles(context.Background(), f.metadata.Path)
	if err != nil {
		return nil, err
	}

	var files []fs.FileInfo

	for _, v := range listFiles {
		if v.Path == f.metadata.Path {
			continue
		}

		files = append(files, fs.FileInfo(v))
	}

	return files, nil
}

func (f *File) Stat() (fs.FileInfo, error) {
	metadata, found := f.datastore.GetMetadataByPath(context.Background(), f.metadata.Path)
	if found != true {
		return nil, nil
	}

	return metadata, nil
}
