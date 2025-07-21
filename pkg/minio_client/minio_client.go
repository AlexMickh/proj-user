package minio_client

import (
	"context"
	"fmt"
	"time"

	"github.com/AlexMickh/proj-user/pkg/utils/retry"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func New(
	ctx context.Context,
	endpoint string,
	user string,
	password string,
	bucketName string,
	isUseSsl bool,
) (*minio.Client, error) {
	const op = "minio-client.New"

	var mc *minio.Client

	err := retry.WithDelay(5, 500*time.Millisecond, func() error {
		var err error

		mc, err = minio.New(endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(user, password, ""),
			Secure: isUseSsl,
		})
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		exists, err := mc.BucketExists(ctx, bucketName)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		if !exists {
			err = mc.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
			if err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return mc, nil
}
