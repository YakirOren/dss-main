package storage

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strconv"

	"cloud.google.com/go/storage"
	"github.com/yakiroren/dss-common/models"
)

func (client Client) ReadFragments(ctx context.Context, fragments []models.Fragment) (io.ReadCloser, error) {
	var readers []io.ReadCloser

	sortedFragments := sortFragments(fragments)
	for _, fragment := range sortedFragments {
		contentReader, err := client.getFragmentContent(ctx, fragment)
		if err != nil {
			return nil, err
		}

		readers = append(readers, contentReader)
	}

	contentMultiReader := CombineReaders(readers...)
	return contentMultiReader, nil
}

func sortFragments(fragments []models.Fragment) []models.Fragment {
	sort.Slice(fragments, func(i, j int) bool {
		a, _ := strconv.Atoi(fragments[i].Name)
		b, _ := strconv.Atoi(fragments[j].Name)

		return a < b
	})

	return fragments
}

func (client Client) getFragmentContent(ctx context.Context, fragment models.Fragment) (*storage.Reader, error) {
	obj := client.gcloud.Bucket(client.bucketName).Object(fmt.Sprintf("attachments/%s/%s/%s", fragment.ChannelID, fragment.MessageID, fragment.Name))
	return obj.NewReader(ctx)
}
