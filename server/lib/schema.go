package lib

import (
	"app/models"
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	twitter "github.com/g8rswimmer/go-twitter/v2"
	"go.uber.org/ratelimit"
)

type Env struct {
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
	pushDataCloseChecks        []bool
	Storage                    interface {
		Setup(ctx context.Context, ErrorLogger *log.Logger) error
		GetCheckpoint() (models.LocalCheckpoint, error)
		PutCheckpoint(checkpoint models.LocalCheckpoint) error
		Put(bucket, object string, data []byte) error
	}
}

type UserRequest struct {
	userID    string
	nextToken string
}

type batchUserRequest struct {
	userIDs []string
}

type FollowMap struct {
	UserID     string
	FollowerID string
}

type PushData struct {
	data string
}

func (env *Env) BatchUserExpander(ctx context.Context, batchUserRequestChan <-chan batchUserRequest) {
	env.RunLogger.Println("in BatchUserExpander")
	defer env.wg.Done()

	for batchUserRequest := range batchUserRequestChan {
		if ok := env.expandUsers(batchUserRequest.userIDs); ok {
			for _, userID := range batchUserRequest.userIDs {
				env.wg.Add(2)
				go env.FollowerExpander(UserRequest{userID: userID, nextToken: ""})
				go env.FriendExpander(UserRequest{userID: userID, nextToken: ""})
			}
		}
		if ctx.Done() != nil {
			close(env.batchUserRequestChan)
		}
	}
}

func (env *Env) FollowerExpander(userRequest UserRequest) {
	env.RunLogger.Println("in FollowerExpander")
	defer env.wg.Done()

	userRequest, done := env.expandFollowers(userRequest)
	if userRequest.nextToken != "" && !done {
		env.wg.Add(1)
		env.FollowerExpander(userRequest)
	}
}

func (env *Env) FriendExpander(userRequest UserRequest) {
	env.RunLogger.Println("in FriendExpander")
	defer env.wg.Done()

	userRequest, done := env.expandFriends(userRequest)
	if userRequest.nextToken != "" && !done {
		env.wg.Add(1)
		go env.FriendExpander(userRequest)
	}
}

func (env *Env) UserDataSaver(ctx context.Context, userDataChan <-chan PushData) {
	env.RunLogger.Println("in UserDataSaver")

	env.saveUserData(ctx, userDataChan, "user_data/")

	env.RunLogger.Println("done in UserDataSaver")
}

func (env *Env) FollowMappingSaver(ctx context.Context, followMapChan <-chan PushData) {
	env.RunLogger.Println("in FollowMappingSaver")

	env.saveUserData(ctx, followMapChan, "follow_map/")

	env.RunLogger.Println("done in FollowMappingSaver")
}

func (env *Env) saveUserData(ctx context.Context, Chan <-chan PushData, filePrefix string) {
	batch := make([]string, 0)
	for obj := range Chan {
		batch = append(batch, obj.data)
		if len(batch) == 1000 {
			file := filePrefix + strconv.Itoa(env.Checkpoint.CurrentEpoc) + "/" + strconv.Itoa(int(time.Now().Unix())) + ".jsonl"
			env.Storage.Put(os.Getenv("BUCKET_NAME"), file, []byte(strings.Join(batch, "\n")))
			batch = make([]string, 0)
		}
		if ctx.Done() != nil {
			break
		}
	}
	if len(batch) > 0 {
		file := filePrefix + strconv.Itoa(env.Checkpoint.CurrentEpoc) + "/" + strconv.Itoa(int(time.Now().Unix())) + ".jsonl"
		env.Storage.Put(os.Getenv("BUCKET_NAME"), file, []byte(strings.Join(batch, "\n")))
	}
}

func (env *Env) Refresh() {
	env.RunLogger.Println("in Refresh")
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	env.batchUserRequestLimiter = ratelimit.New(1)
	env.userFollowerRequestLimiter = ratelimit.New(1)
	env.userFriendRequestLimiter = ratelimit.New(1)

	// setup env channels
	env.batchUserRequestChan = make(chan batchUserRequest, 360)
	env.userDataChan = make(chan PushData, 100000)
	env.followMapChan = make(chan PushData, 100000)

	env.pushDataCloseChecks = make([]bool, 2)

	// setup env waitgroups
	env.wg = &sync.WaitGroup{}

	// start workers
	env.wg.Add(1)
	go env.BatchUserExpander(ctx, env.batchUserRequestChan)

	batch := make([]string, 0)
	for userID, userStub := range env.Checkpoint.UserMap {
		if userStub.InNextEpoc {
			batch = append(batch, userID)
			if len(batch) == 1000 {
				env.batchUserRequestChan <- batchUserRequest{
					userIDs: batch,
				}
				batch = make([]string, 0)
			}
		}
	}
	if len(batch) > 0 {
		env.batchUserRequestChan <- batchUserRequest{
			userIDs: batch,
		}
	}

	// start closing cascade
	cancel()

	// wait for all workers to finish
	env.wg.Wait()

	// signal data recording functions to stop receiving data
	close(env.followMapChan)
	close(env.userDataChan)

	env.RunLogger.Println("done Refresh")
}
