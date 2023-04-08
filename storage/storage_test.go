package storage

import (
	"github.com/yakiroren/dss-common/models"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_sortFragments(t *testing.T) {
	require.Equal(t,
		sortFragments(unsortedFragments()),
		sortedFragments())
}

func sortedFragments() []models.Fragment {
	return []models.Fragment{
		{
			Name:      "1",
			MessageID: "",
			ChannelID: "",
			Size:      0,
		},
		{
			Name:      "2",
			MessageID: "",
			ChannelID: "",
			Size:      0,
		},
	}
}

func unsortedFragments() []models.Fragment {
	return []models.Fragment{
		{
			Name:      "2",
			MessageID: "",
			ChannelID: "",
			Size:      0,
		},
		{
			Name:      "1",
			MessageID: "",
			ChannelID: "",
			Size:      0,
		},
	}
}
