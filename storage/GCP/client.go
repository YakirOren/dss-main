package GCP

import (
	"context"

	"cloud.google.com/go/storage"

	"google.golang.org/api/option"
)

const bucketName = "discord"

type Client struct {
	gcloud     *storage.Client
	bucketName string
}

func NewClient() (*Client, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithoutAuthentication())
	if err != nil {
		return nil, err
	}

	return &Client{gcloud: client, bucketName: bucketName}, nil
}
