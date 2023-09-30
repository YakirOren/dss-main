package GCP

import (
	"cloud.google.com/go/storage"
	"context"
	ds "dss-main/storage"
	"fmt"
	"github.com/yakiroren/dss-common/models"
	"io"
)

func (client Client) ReadFragments(ctx context.Context, fragments []models.Fragment) (io.ReadCloser, error) {
	var readers []io.ReadCloser

	sortedFragments := ds.SortFragments(fragments)
	for _, fragment := range sortedFragments {
		contentReader, err := client.getFragmentContent(ctx, fragment)
		if err != nil {
			return nil, err
		}

		readers = append(readers, contentReader)
	}

	contentMultiReader := ds.CombineReaders(readers...)
	return contentMultiReader, nil
}

func (client Client) getFragmentContent(ctx context.Context, fragment models.Fragment) (*storage.Reader, error) {
	obj := client.gcloud.Bucket(client.bucketName)
	bucket := obj.Object(fmt.Sprintf("attachments/%s/%s/%s", fragment.ChannelID, fragment.MessageID, fragment.Name))
	return bucket.NewReader(ctx)
}
