package ds

import (
	"context"
	"github.com/yakiroren/dss-common/models"
	"io"
	"sort"
	"strconv"
)

type Client interface {
	ReadFragments(ctx context.Context, fragments []models.Fragment) (io.ReadCloser, error)
}

func SortFragments(fragments []models.Fragment) []models.Fragment {
	sort.Slice(fragments, func(i, j int) bool {
		a, _ := strconv.Atoi(fragments[i].Name)
		b, _ := strconv.Atoi(fragments[j].Name)

		return a < b
	})

	return fragments
}
