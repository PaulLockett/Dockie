package lib

import (
	"app/models"
	"context"
	"log"
	"sync"
	"testing"
	"time"

	twitter "github.com/g8rswimmer/go-twitter/v2"
	"go.uber.org/ratelimit"
)

func TestEnv_Refresh(t *testing.T) {
	type fields struct {
		StartTime                  time.Time
		Checkpoint                 models.LocalCheckpoint
		RunLogger                  *log.Logger
		ErrorLogger                *log.Logger
		ApiKey                     string
		wg                         *sync.WaitGroup
		TwitterClient              *twitter.Client
		userDataChan               chan PushData
		batchUserRequestChan       chan batchUserRequest
		batchUserRequestLimiter    ratelimit.Limiter
		userFollowerRequestLimiter ratelimit.Limiter
		userFriendRequestLimiter   ratelimit.Limiter
		followMapChan              chan PushData
		Storage                    interface {
			Setup(ctx context.Context, ErrorLogger *log.Logger) error
			GetCheckpoint() (models.LocalCheckpoint, error)
			PutCheckpoint(checkpoint models.LocalCheckpoint) error
			Put(bucket, object string, data []byte) error
		}
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := &Env{
				StartTime:                  tt.fields.StartTime,
				Checkpoint:                 tt.fields.Checkpoint,
				RunLogger:                  tt.fields.RunLogger,
				ErrorLogger:                tt.fields.ErrorLogger,
				ApiKey:                     tt.fields.ApiKey,
				wg:                         tt.fields.wg,
				TwitterClient:              tt.fields.TwitterClient,
				userDataChan:               tt.fields.userDataChan,
				batchUserRequestChan:       tt.fields.batchUserRequestChan,
				batchUserRequestLimiter:    tt.fields.batchUserRequestLimiter,
				userFollowerRequestLimiter: tt.fields.userFollowerRequestLimiter,
				userFriendRequestLimiter:   tt.fields.userFriendRequestLimiter,
				followMapChan:              tt.fields.followMapChan,
				Storage:                    tt.fields.Storage,
			}
			env.Refresh()
		})
	}
}

func TestEnv_saveUserData(t *testing.T) {
	type fields struct {
		StartTime                  time.Time
		Checkpoint                 models.LocalCheckpoint
		RunLogger                  *log.Logger
		ErrorLogger                *log.Logger
		ApiKey                     string
		wg                         *sync.WaitGroup
		TwitterClient              *twitter.Client
		userDataChan               chan PushData
		batchUserRequestChan       chan batchUserRequest
		batchUserRequestLimiter    ratelimit.Limiter
		userFollowerRequestLimiter ratelimit.Limiter
		userFriendRequestLimiter   ratelimit.Limiter
		followMapChan              chan PushData
		Storage                    interface {
			Setup(ctx context.Context, ErrorLogger *log.Logger) error
			GetCheckpoint() (models.LocalCheckpoint, error)
			PutCheckpoint(checkpoint models.LocalCheckpoint) error
			Put(bucket, object string, data []byte) error
		}
	}
	type args struct {
		Chan       <-chan PushData
		filePrefix string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := &Env{
				StartTime:                  tt.fields.StartTime,
				Checkpoint:                 tt.fields.Checkpoint,
				RunLogger:                  tt.fields.RunLogger,
				ErrorLogger:                tt.fields.ErrorLogger,
				ApiKey:                     tt.fields.ApiKey,
				wg:                         tt.fields.wg,
				TwitterClient:              tt.fields.TwitterClient,
				userDataChan:               tt.fields.userDataChan,
				batchUserRequestChan:       tt.fields.batchUserRequestChan,
				batchUserRequestLimiter:    tt.fields.batchUserRequestLimiter,
				userFollowerRequestLimiter: tt.fields.userFollowerRequestLimiter,
				userFriendRequestLimiter:   tt.fields.userFriendRequestLimiter,
				followMapChan:              tt.fields.followMapChan,
				Storage:                    tt.fields.Storage,
			}
			env.saveUserData(tt.args.Chan, tt.args.filePrefix)
		})
	}
}
