package lib

import (
	"app/models"
	"context"
	"log"
	"sync"
	"time"

	twitter "github.com/g8rswimmer/go-twitter/v2"
	"go.uber.org/ratelimit"
)

// Env is active the environment for the application
type Env struct {
	StartTime                  time.Time
	Checkpoint                 models.LocalCheckpoint
	APIKey                     string
	wg                         *sync.WaitGroup
	TwitterClient              *twitter.Client
	dataChan                   chan any
	batchUserRequestChan       chan batchUserRequest
	batchUserRequestLimiter    ratelimit.Limiter
	userFollowerRequestLimiter ratelimit.Limiter
	userFriendRequestLimiter   ratelimit.Limiter
	Storage                    interface {
		Setup(url string, user string, password string, nameSpace string, database string) error
		GetCheckpoint() (models.LocalCheckpoint, error)
		PutCheckpoint(checkpoint models.LocalCheckpoint) error
		Put(name string, data interface{}) error
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

// BatchUserExpander pulls the user data for a set of user Ids and sends each user id to the friends and followers expanders
func (env *Env) BatchUserExpander(ctx context.Context, batchUserRequestChan <-chan batchUserRequest) {
	log.Println("in BatchUserExpander")
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
	log.Println("in FollowerExpander")
	defer env.wg.Done()

	userRequest, done := env.expandFollowers(userRequest)
	if userRequest.nextToken != "" && !done {
		env.wg.Add(1)
		env.FollowerExpander(userRequest)
	}
}

// FriendExpander pulls the user data of all a userid's friends (users they follow) and sends it to be recorded
func (env *Env) FriendExpander(userRequest UserRequest) {
	log.Println("in FriendExpander")
	defer env.wg.Done()

	userRequest, done := env.expandFriends(userRequest)
	if userRequest.nextToken != "" && !done {
		env.wg.Add(1)
		env.FriendExpander(userRequest)
	}
}

// DataSaver batches data packets in it's channel and writes them to cloud storage
func (env *Env) DataSaver(Chan <-chan any) {
	for obj := range Chan {
		tableName := "error"
		switch obj := obj.(type) {
		case twitter.UserObj:
			userID := obj.ID
			tableName = "user_data" + ":" + userID
		case FollowMap:
			userID := obj.UserID
			FollowerID := obj.FollowerID
			tableName = "follow_map" + ":`" + userID + "-" + FollowerID + "`"
		default:
			log.Println("unknown type in saveUserData")
		}
		if err := env.Storage.Put(tableName, obj); err != nil {
			log.Println(err)
		}
	}
}

// Refresh pulls the user data for each userid, their follower and their friends (people they follow) and records them on the cloud
func (env *Env) Refresh() {
	log.Println("in Refresh")
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	env.batchUserRequestLimiter = ratelimit.New(1)
	env.userFollowerRequestLimiter = ratelimit.New(1)
	env.userFriendRequestLimiter = ratelimit.New(1)

	// setup env channels
	env.batchUserRequestChan = make(chan batchUserRequest, 360)
	env.dataChan = make(chan any, 100000)

	// setup data output function
	go env.DataSaver(env.dataChan)

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
				log.Println("sending batch of users")
				env.batchUserRequestChan <- batchUserRequest{
					userIDs: batch,
				}
				log.Println("sent batch of users")
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
	log.Println("closing batchUserRequestChan")
	cancel()
	log.Println("closed batchUserRequestChan")

	// wait for all workers to finish
	log.Println("waiting for workers to finish")
	env.wg.Wait()
	log.Println("workers finished")

	// signal data recording functions to stop receiving data
	log.Println("closing data channels")
	close(env.dataChan)
	log.Println("closed data channels")

	log.Println("done Refresh")
}
