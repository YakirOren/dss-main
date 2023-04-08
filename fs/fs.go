package fs

import (
	"context"
	"dss-main/storage"
	"errors"
	"github.com/yakiroren/dss-common/db"
	"net/http"
	"strings"
)

type FS struct {
	storage   *storage.Client
	datastore db.DataStore
}

func (fs FS) Open(path string) (http.File, error) {
	if path != "/" {
		path = strings.TrimSuffix(path, "/")
	}

	metadata, found := fs.datastore.GetMetadataByPath(context.Background(), path)
	if found != true {
		return nil, errors.New("file not found")
	}

	return &File{
		metadata:  metadata,
		storage:   fs.storage,
		datastore: fs.datastore}, nil
}

func New(store db.DataStore) (*FS, error) {
	storageClient, err := storage.NewClient()
	if err != nil {
		return nil, err
	}

	return &FS{datastore: store, storage: storageClient}, nil
}
