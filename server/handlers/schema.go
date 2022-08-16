package handlers

import (
	"app/models"
	"context"
	"time"
)

type Env struct {
	StartTime time.Time
	Storage   interface {
		Setup(ctx context.Context) error
		Get(ctx context.Context, bucket, object string) ([]byte, error)
		Put(ctx context.Context, bucket, object string, data []byte) error
		GetCheckpoint(ctx context.Context, bucket string) (models.LocalCheckpoint, error)
		PutCheckpoint(ctx context.Context, bucket string, checkpoint models.LocalCheckpoint) error
	}
}
