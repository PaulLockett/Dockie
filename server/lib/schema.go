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
	ErrorLogger          *log.Logger
	ApiKey               string
	wg                   *sync.WaitGroup
	TwitterClient        *twitter.Client
	userFollowerChan     chan userRequest
	userFriendChan       chan userRequest
	userDataChan         chan string
	batchUserRequestChan chan batchUserRequest
	followMapChan        chan string
	Storage              interface {
		Setup(ctx context.Context, ErrorLogger *log.Logger) error
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

	for userRequest := range userFollowerRequestChan {
		if !userRequest.rateLimitReset.IsZero() {
			env.RunLogger.Printf("sleeping for %s in FollowerExpander", time.Until(userRequest.rateLimitReset))
			time.Sleep(time.Until(userRequest.rateLimitReset))
			env.RunLogger.Printf("done sleeping in FollowerExpander")
		}
		env.expandFollowers(userRequest.userID, userRequest.nextToken)
	}
}

func (env *Env) FriendExpander(userFriendRequestChan <-chan userRequest) {
	env.RunLogger.Println("in FriendExpander")

	for userRequest := range userFriendRequestChan {
		if !userRequest.rateLimitReset.IsZero() {
			env.RunLogger.Printf("sleeping for %s in FriendExpander", time.Until(userRequest.rateLimitReset))
			time.Sleep(time.Until(userRequest.rateLimitReset))
			env.RunLogger.Printf("done sleeping in FriendExpander")
		}
		env.expandFriends(userRequest.userID, userRequest.nextToken)

	}
}

func (env *Env) BatchUserExpander(batchUserRequestChan <-chan batchUserRequest) {
	env.RunLogger.Println("in BatchUserExpander")

	for batchUserRequest := range batchUserRequestChan {
		if !batchUserRequest.rateLimitReset.IsZero() {
			env.RunLogger.Printf("sleeping for %s in BatchUserExpander", time.Until(batchUserRequest.rateLimitReset))
			time.Sleep(time.Until(batchUserRequest.rateLimitReset))
			env.RunLogger.Printf("done sleeping in BatchUserExpander")
		}
		env.expandUsers(batchUserRequest.userIDs)
	}
}

func (env *Env) UserDataSaver(userDataChan <-chan string) {
	env.RunLogger.Println("in UserDataSaver")
	batch := make([]string, 0)
	for obj := range userDataChan {
		batch = append(batch, obj)
		if len(batch) == 1000 {
			file := "user_data/" + strconv.Itoa(env.Checkpoint.CurrentEpoc) + "/" + strconv.Itoa(int(time.Now().Unix())) + ".jsonl"
			env.Storage.Put(os.Getenv("BUCKET_NAME"), file, []byte(strings.Join(batch, "\n")))
			batch = make([]string, 0)
		}
	}
	if len(batch) > 0 {
		file := "user_data/" + strconv.Itoa(env.Checkpoint.CurrentEpoc) + "/" + strconv.Itoa(int(time.Now().Unix())) + ".jsonl"
		env.Storage.Put(os.Getenv("BUCKET_NAME"), file, []byte(strings.Join(batch, "\n")))
	}
	env.RunLogger.Println("done in UserDataSaver")
}

func (env *Env) FollowMappingSaver(followMapChan <-chan string) {
	env.RunLogger.Println("in FollowMappingSaver")
	batch := make([]string, 0)

	for obj := range followMapChan {
		batch = append(batch, obj)
		if len(batch) == 1000 {
			file := "follow_map/" + strconv.Itoa(env.Checkpoint.CurrentEpoc) + "/" + strconv.Itoa(int(time.Now().Unix())) + ".jsonl"
			env.Storage.Put(os.Getenv("BUCKET_NAME"), file, []byte(strings.Join(batch, "\n")))
			batch = make([]string, 0)
		}
	}
	if len(batch) > 0 {
		file := "follow_map/" + strconv.Itoa(env.Checkpoint.CurrentEpoc) + "/" + strconv.Itoa(int(time.Now().Unix())) + ".jsonl"
		env.Storage.Put(os.Getenv("BUCKET_NAME"), file, []byte(strings.Join(batch, "\n")))
	}
	env.RunLogger.Println("done in FollowMappingSaver")
}

func (env *Env) Refresh() {
	env.RunLogger.Println("in Refresh")
	// setup env channels
	// 360 is max number of Twitter API calls per 15 minutes per 24 hour period per Authenticated User
	env.userFollowerChan = make(chan userRequest, 360)
	env.userFriendChan = make(chan userRequest, 360)
	env.batchUserRequestChan = make(chan batchUserRequest, 360)

	env.userDataChan = make(chan string, 10000)
	env.followMapChan = make(chan string, 10000)

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

	env.RunLogger.Println("done Refresh")
}
