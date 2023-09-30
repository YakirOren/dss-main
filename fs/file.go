package fs

import (
	"context"
	ds "dss-main/storage"
	"errors"
	"io"
	"io/fs"
	"path/filepath"

	"github.com/yakiroren/dss-common/db"
	"github.com/yakiroren/dss-common/models"
)

type File struct {
	reader    io.ReadCloser
	metadata  *models.FileMetadata
	datastore db.DataStore
	storage   ds.Client
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

func (f File) Seek(_ int64, _ int) (int64, error) {
	return 0, nil
}

func (f File) Readdir(_ int) ([]fs.FileInfo, error) {
	if !f.metadata.IsDir() {
		return nil, errors.New("file is not a directory")
	}

	listFiles, err := f.datastore.ListFiles(context.Background(), filepath.Join(f.metadata.Path, f.metadata.FileName))
	if err != nil {
		return nil, err
	}

	var files []fs.FileInfo

	for _, v := range listFiles {
		// don't return itself
		if v.Name() == f.metadata.Path {
			continue
		}

		files = append(files, fs.FileInfo(v))
	}

	return files, nil
}

func (f *File) Stat() (fs.FileInfo, error) {
	metadata, found := f.datastore.GetMetadataByID(context.Background(), f.metadata.Id.Hex())
	if !found {
		return nil, nil
	}

	return metadata, nil
}
