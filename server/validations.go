package server

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

func (s *Server) fixFilename(ctx context.Context, filename string, path string) string {
	filename = sanitizeFilename(filename)
	ext := filepath.Ext(filename)
	newFilename := filename

	i := 1
	for {
		_, found := s.datastore.GetMetadataByPath(ctx, fmt.Sprintf("%s/%s", path, newFilename))
		if !found {
			return newFilename
		}

		newFilename = fileNameWithoutExtTrimSuffix(filename) + fmt.Sprintf("(%d)", i) + ext
		i++
	}
}

func fileNameWithoutExtTrimSuffix(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

func sanitizeFilename(filename string) string {
	// Define a regular expression to match against the filename
	pattern := "[^a-zA-Z0-9_.-]+"
	regex := regexp.MustCompile(pattern)

	// Replace all spaces with underscores
	filename = strings.ReplaceAll(filename, " ", "_")

	// Replace all non-matching characters with an empty string
	sanitized := regex.ReplaceAllString(filename, "")

	return sanitized
}

func validatePath(path string) bool {
	// Regular expression pattern for validating Linux paths
	pattern := `^/([a-zA-Z0-9_-]+/?)*$`

	// Compile the regular expression pattern
	regex := regexp.MustCompile(pattern)

	// Use the MatchString function to check if the path matches the pattern
	return regex.MatchString(path)
}

func getPathSegments(path string) []string {
	// Remove trailing slashes from the path
	path = strings.TrimSuffix(path, "/")

	// Split the path into its directory segments
	segments := strings.Split(path, "/")

	// Reverse the order of the segments
	for i, j := 0, len(segments)-1; i < j; i, j = i+1, j-1 {
		segments[i], segments[j] = segments[j], segments[i]
	}

	// Build an array of paths from root to deepest directory
	var paths []string
	currentPath := ""
	for i := len(segments) - 1; i >= 0; i-- {
		currentPath += segments[i] + "/"
		paths = append(paths, currentPath)
	}

	return paths
}
