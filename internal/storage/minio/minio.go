package minio

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
)

type Minio struct {
	mc         *minio.Client
	bucketName string
}

const defaultImage = "avatar.png"

func New(mc *minio.Client, bucketName string) *Minio {
	return &Minio{
		mc:         mc,
		bucketName: bucketName,
	}
}

func (m *Minio) SaveAvatar(ctx context.Context, id string, avatar []byte) (string, error) {
	const op = "storage.minio.user.SaveAvatar"

	if avatar == nil {
		url, err := m.GetImageUrl(ctx, defaultImage)
		if err != nil {
			return "", fmt.Errorf("%s: %w", op, err)
		}

		return url, nil
	}

	reader := bytes.NewReader(avatar)

	_, err := m.mc.PutObject(
		ctx,
		m.bucketName,
		id,
		reader,
		int64(len(avatar)),
		minio.PutObjectOptions{ContentType: "image/png"},
	)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	url, err := m.GetImageUrl(ctx, id)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return url, nil
}

func (m *Minio) GetImageUrl(ctx context.Context, avatarId string) (string, error) {
	const op = "storage.minio.GetImage"

	url, err := m.mc.PresignedGetObject(ctx, m.bucketName, avatarId, 5*24*time.Hour, nil)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return url.String(), nil
}
