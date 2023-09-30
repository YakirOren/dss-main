package Discord

import (
	"context"
	ds "dss-main/storage"
	"github.com/yakiroren/dss-common/models"
	"io"
	"net/http"
	"path/filepath"
)

const CdnPrefix = "cdn.discordapp.com/attachments"

type Client struct {
}

func (client Client) ReadFragments(ctx context.Context, fragments []models.Fragment) (io.ReadCloser, error) {
	var readers []io.ReadCloser

	sortedFragments := ds.SortFragments(fragments)
	for _, fragment := range sortedFragments {
		url := client.getPath(fragment)

		response, err := http.Get("https://" + url)
		if err != nil {
			return nil, err
		}

		readers = append(readers, response.Body)
	}

	return ds.CombineReaders(readers...), nil
}

func (client Client) getPath(fragment models.Fragment) string {
	return filepath.Join(CdnPrefix, fragment.ChannelID, fragment.MessageID, fragment.Name)
}
