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

// BatchUserExpander pulls the user data for a set of user Ids and sends each user id to the friends and followers expanders
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

// FollowerExpander pulls the user data of all a userid's followers and sends it to be recorded
func (env *Env) FollowerExpander(userRequest UserRequest) {
	env.RunLogger.Println("in FollowerExpander")
	defer env.wg.Done()

	userRequest, done := env.expandFollowers(userRequest)
	if userRequest.nextToken != "" && !done {
		env.wg.Add(1)
		env.FollowerExpander(userRequest)
	}
}

// FriendExpander pulls the user data of all a userid's friends (users they follow) and sends it to be recorded
func (env *Env) FriendExpander(userRequest UserRequest) {
	env.RunLogger.Println("in FriendExpander")
	defer env.wg.Done()

	userRequest, done := env.expandFriends(userRequest)
	if userRequest.nextToken != "" && !done {
		env.wg.Add(1)
		env.FriendExpander(userRequest)
	}
}

// UserDataSaver wraps a saveUserData implementation for userdata objects
func (env *Env) UserDataSaver(userDataChan <-chan PushData) {
	env.RunLogger.Println("in UserDataSaver")

	env.saveUserData(userDataChan, "user_data/")

	env.RunLogger.Println("done in UserDataSaver")
}

// FollowMappingSaver wraps a saveUserData implementation for follower mapping objects
func (env *Env) FollowMappingSaver(followMapChan <-chan PushData) {
	env.RunLogger.Println("in FollowMappingSaver")

	env.saveUserData(followMapChan, "follow_map/")

	env.RunLogger.Println("done in FollowMappingSaver")
}

// saveUserData batches data packets in it's channel and writes them to cloud storage
func (env *Env) saveUserData(Chan <-chan PushData, filePrefix string) {
	batch := make([]string, 0)
	for obj := range Chan {
		batch = append(batch, obj.data)
		if len(batch) == 1000 {
			env.RunLogger.Println("writing batch")
			file := filePrefix + strconv.Itoa(env.Checkpoint.CurrentEpoc) + "/" + strconv.Itoa(int(time.Now().Unix())) + ".jsonl"
			env.Storage.Put(os.Getenv("BUCKET_NAME"), file, []byte(strings.Join(batch, "\n")))
			env.RunLogger.Println("wrote batch")
			batch = make([]string, 0)
		}
	}
	if len(batch) > 0 {
		file := filePrefix + strconv.Itoa(env.Checkpoint.CurrentEpoc) + "/" + strconv.Itoa(int(time.Now().Unix())) + ".jsonl"
		env.Storage.Put(os.Getenv("BUCKET_NAME"), file, []byte(strings.Join(batch, "\n")))
	}
}

// Refresh pulls the user data for each userid, their follower and their friends (people they follow) and records them on the cloud
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

	// setup data output function
	go env.UserDataSaver(env.userDataChan)
	go env.FollowMappingSaver(env.followMapChan)

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
				env.RunLogger.Println("sending batch of users")
				env.batchUserRequestChan <- batchUserRequest{
					userIDs: batch,
				}
				env.RunLogger.Println("sent batch of users")
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
	env.RunLogger.Println("closing batchUserRequestChan")
	cancel()
	env.RunLogger.Println("closed batchUserRequestChan")

	// wait for all workers to finish
	env.RunLogger.Println("waiting for workers to finish")
	env.wg.Wait()
	env.RunLogger.Println("workers finished")

	// signal data recording functions to stop receiving data
	env.RunLogger.Println("closing data channels")
	close(env.followMapChan)
	close(env.userDataChan)
	env.RunLogger.Println("closed data channels")

	env.RunLogger.Println("done Refresh")
}
