// Package storage wraps the MinIO object store admin uploads write to. It
// is the write side of apps/web/src/routes/api/media/$.ts, which already
// reads from the same buckets (lesson-audio, stories, toeic, ielts).
package storage

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"engflex-api/config"
)

// Client uploads objects to a MinIO bucket. Buckets are provisioned ahead
// of time (docker-compose's minio-init) with anonymous download already
// set — Client only ever writes, never creates or configures a bucket.
type Client struct {
	mc *minio.Client
}

func NewClient(cfg config.MinioConfig) (*Client, error) {
	endpoint, secure := splitEndpoint(cfg.Endpoint)
	mc, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: secure || cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}
	return &Client{mc: mc}, nil
}

// splitEndpoint strips the scheme minio-go doesn't want in its host arg,
// reporting whether the scheme implies TLS.
func splitEndpoint(raw string) (host string, secure bool) {
	switch {
	case len(raw) >= 8 && raw[:8] == "https://":
		return raw[8:], true
	case len(raw) >= 7 && raw[:7] == "http://":
		return raw[7:], false
	default:
		return raw, false
	}
}

// Upload writes one object, overwriting any existing object at the same
// key. Returns the key unchanged — callers map it to a public URL
// themselves (see stories' /api/media/stories/{key} convention).
func (c *Client) Upload(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) error {
	_, err := c.mc.PutObject(ctx, bucket, key, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}
