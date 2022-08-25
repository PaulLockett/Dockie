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
)

type Env struct {
	StartTime            time.Time
	Checkpoint           models.LocalCheckpoint
	RunLogger            *log.Logger
	ApiKey               string
	wg                   *sync.WaitGroup
	TwitterClient        *twitter.Client
	userFollowerChan     chan userRequest
	userFriendChan       chan userRequest
	userDataChan         chan string
	batchUserRequestChan chan batchUserRequest
	followMapChan        chan string
	Storage              interface {
		Setup(ctx context.Context) error
		GetCheckpoint() (models.LocalCheckpoint, error)
		PutCheckpoint(checkpoint models.LocalCheckpoint) error
		Put(bucket, object string, data []byte) error
	}
}

type userRequest struct {
	userID         string
	nextToken      string
	rateLimitReset time.Time
}

type batchUserRequest struct {
	userIDs        []string
	rateLimitReset time.Time
}

type FollowMap struct {
	UserID     string
	FollowerID string
}

func (env *Env) FollowerExpander(userFollowerRequestChan <-chan userRequest) {
	env.RunLogger.Println("in FollowerExpander")
	defer env.wg.Done()

	for userRequest := range userFollowerRequestChan {
		if !userRequest.rateLimitReset.IsZero() {
			time.Sleep(time.Until(userRequest.rateLimitReset))
		}
		env.expandFollowers(userRequest.userID, userRequest.nextToken)
	}
}

func (env *Env) FriendExpander(userFriendRequestChan <-chan userRequest) {
	env.RunLogger.Println("in FriendExpander")
	defer env.wg.Done()

	for userRequest := range userFriendRequestChan {
		if !userRequest.rateLimitReset.IsZero() {
			time.Sleep(time.Until(userRequest.rateLimitReset))
		}
		env.expandFriends(userRequest.userID, userRequest.nextToken)

	}
}

func (env *Env) BatchUserExpander(batchUserRequestChan <-chan batchUserRequest) {
	env.RunLogger.Println("in BatchUserExpander")
	defer env.wg.Done()

	for batchUserRequest := range batchUserRequestChan {
		if !batchUserRequest.rateLimitReset.IsZero() {
			time.Sleep(time.Until(batchUserRequest.rateLimitReset))
		}
		env.expandUsers(batchUserRequest.userIDs)
	}
}

func (env *Env) UserDataSaver(userDataChan <-chan string) {
	env.RunLogger.Println("in UserDataSaver")
	defer env.wg.Done()
	env.putData(userDataChan, "user_data/")
}

func (env *Env) FollowMappingSaver(followMapChan <-chan string) {
	env.RunLogger.Println("in FollowMappingSaver")
	defer env.wg.Done()
	env.putData(followMapChan, "follow_map/")
}

func (env *Env) putData(Chan <-chan string, directoryPrefix string) {
	env.RunLogger.Println("in putData")
	batch := make([]string, 0)

	for obj := range Chan {
		batch = append(batch, obj)
		if len(batch) == 1000 {
			file := directoryPrefix + strconv.Itoa(env.Checkpoint.CurrentEpoc) + "/" + time.Now().String() + ".jsonl"
			go env.Storage.Put(os.Getenv("BUCKET_NAME"), file, []byte(strings.Join(batch, "\n")))
			batch = make([]string, 0)
		}
	}
	if len(batch) > 0 {
		file := directoryPrefix + strconv.Itoa(env.Checkpoint.CurrentEpoc) + "/" + time.Now().String() + ".jsonl"
		go env.Storage.Put(os.Getenv("BUCKET_NAME"), file, []byte(strings.Join(batch, "\n")))
	}
}

func (env *Env) Refresh() {
	env.RunLogger.Println("in Refresh")
	// setup env channels
	// 360 is max number of Twitter API calls per 15 minutes per 24 hour period per Authenticated User
	env.userFollowerChan = make(chan userRequest, 360)
	env.userFriendChan = make(chan userRequest, 360)
	env.batchUserRequestChan = make(chan batchUserRequest, 360)

	env.userDataChan = make(chan string)
	env.followMapChan = make(chan string)

	env.wg = &sync.WaitGroup{}
	env.wg.Add(5)
	// start expander workers
	go env.FollowerExpander(env.userFollowerChan)
	go env.FriendExpander(env.userFriendChan)
	go env.BatchUserExpander(env.batchUserRequestChan)

	go env.UserDataSaver(env.userDataChan)
	go env.FollowMappingSaver(env.followMapChan)

	batch := make([]string, 0)
	for userID, userStub := range env.Checkpoint.UserMap {
		if userStub.InNextEpoc {
			batch = append(batch, userID)
			if len(batch) == 1000 {
				env.batchUserRequestChan <- batchUserRequest{
					userIDs:        batch,
					rateLimitReset: time.Time{},
				}
				batch = make([]string, 0)
			}
			env.userFollowerChan <- userRequest{
				userID:         userID,
				nextToken:      "",
				rateLimitReset: time.Time{},
			}
			env.userFriendChan <- userRequest{
				userID:         userID,
				nextToken:      "",
				rateLimitReset: time.Time{},
			}
		}
	}
	if len(batch) > 0 {
		env.batchUserRequestChan <- batchUserRequest{
			userIDs:        batch,
			rateLimitReset: time.Time{},
		}
	}

	env.wg.Wait()
}
