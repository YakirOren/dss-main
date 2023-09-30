package fs

import (
	"context"
	ds "dss-main/storage"
	discord "dss-main/storage/Discord"
	"errors"
	"net/http"
	"strings"

	"github.com/yakiroren/dss-common/db"
)

type FS struct {
	storage   ds.Client
	datastore db.DataStore
}

func (fs FS) Open(path string) (http.File, error) {
	if path != "/" {
		path = strings.TrimSuffix(path, "/")
	}

	metadata, found := fs.datastore.GetMetadataByPath(context.Background(), path)
	if !found {
		return nil, errors.New("file not found")
	}

	return &File{
		metadata:  metadata,
		storage:   fs.storage,
		datastore: fs.datastore,
	}, nil
}

func New(store db.DataStore) (*FS, error) {
	return &FS{datastore: store, storage: discord.Client{}}, nil
}
