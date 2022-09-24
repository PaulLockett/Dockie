package lib

import (
	"app/models"
	"context"
	"log"
	"reflect"
	"sync"
	"testing"
	"time"

	twitter "github.com/g8rswimmer/go-twitter/v2"
	"go.uber.org/ratelimit"
)

func TestEnv_expandFollowers(t *testing.T) {
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
		user UserRequest
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   UserRequest
		want1  bool
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
			got, got1 := env.expandFollowers(tt.args.user)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Env.expandFollowers() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Env.expandFollowers() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestEnv_expandFriends(t *testing.T) {
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
		user UserRequest
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   UserRequest
		want1  bool
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
			got, got1 := env.expandFriends(tt.args.user)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Env.expandFriends() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Env.expandFriends() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestEnv_expandUsers(t *testing.T) {
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
		userIDs []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
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
			if got := env.expandUsers(tt.args.userIDs); got != tt.want {
				t.Errorf("Env.expandUsers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnv_reportError(t *testing.T) {
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
		userId        string
		err           error
		partialErrors []*twitter.ErrorObj
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
			env.reportError(tt.args.userId, tt.args.err, tt.args.partialErrors)
		})
	}
}

func TestEnv_sendUserData(t *testing.T) {
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
		dictionaries map[string]*twitter.UserDictionary
		keepFresh    bool
		mappings     *[]FollowMap
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
			env.sendUserData(tt.args.dictionaries, tt.args.keepFresh, tt.args.mappings)
		})
	}
}
