package storage

import (
	"io"
)

type CombinedReadCloser struct {
	readers []io.ReadCloser // the input readers
	multi   io.Reader       // the combined reader
}

func CombineReaders(readers ...io.ReadCloser) io.ReadCloser {
	// create a new CombinedReadCloser instance
	var newReaders []io.Reader
	for _, reader := range readers {
		newReaders = append(newReaders, reader)
	}

	return &CombinedReadCloser{
		readers: readers,
		multi:   io.MultiReader(newReaders...),
	}
}

func (r CombinedReadCloser) Read(p []byte) (int, error) {
	// read from the combined reader
	return r.multi.Read(p)
}

func (r CombinedReadCloser) Close() error {
	// close each input reader
	for _, reader := range r.readers {
		err := reader.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
