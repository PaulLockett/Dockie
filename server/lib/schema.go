package lib

import (
	"app/models"
	"context"
	"time"
)

type Env struct {
	StartTime  time.Time
	Checkpoint models.LocalCheckpoint
	Storage    interface {
		Setup(ctx context.Context) error
		GetCheckpoint() (models.LocalCheckpoint, error)
		PutCheckpoint(checkpoint models.LocalCheckpoint) error
	}
}
