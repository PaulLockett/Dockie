package lib

import (
	"app/models"
	"context"
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
	wg                   *sync.WaitGroup
	TwitterClient        *twitter.Client
	userIDChan           chan userRequest
	userDataChan         chan string
	batchUserRequestChan chan batchUserRequest
	followMapChan        chan followMap
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

type followMap struct {
	userID     string
	followerID string
}

func (env *Env) FollowerExpander(userRequestChan <-chan userRequest) {
	defer env.wg.Done()

	for userRequest := range userRequestChan {
		if !userRequest.rateLimitReset.IsZero() {
			time.Sleep(time.Until(userRequest.rateLimitReset))
		}
		env.expandFollowers(userRequest.userID, userRequest.nextToken)
	}
}

func (env *Env) FriendExpander(userRequestChan <-chan userRequest) {
	defer env.wg.Done()

	for userRequest := range userRequestChan {
		if !userRequest.rateLimitReset.IsZero() {
			time.Sleep(time.Until(userRequest.rateLimitReset))
		}
		env.expandFriends(userRequest.userID, userRequest.nextToken)

	}
}

func (env *Env) BatchUserExpander(batchUserRequestChan <-chan batchUserRequest) {
	defer env.wg.Done()

	for batchUserRequest := range batchUserRequestChan {
		if !batchUserRequest.rateLimitReset.IsZero() {
			time.Sleep(time.Until(batchUserRequest.rateLimitReset))
		}
		env.expandUsers(batchUserRequest.userIDs)
	}
}

func (env *Env) UserDataSaver(userDataChan <-chan string) {
	defer env.wg.Done()
	env.putData(userDataChan, "user_data_")
}

func (env *Env) FollowMappingSaver(followMapChan <-chan string) {
	defer env.wg.Done()
	env.putData(followMapChan, "follow_map_")
}

func (env *Env) putData(Chan <-chan string, directoryPrefix string) {
	batch := make([]string, 0)

	for obj := range Chan {
		batch = append(batch, obj)
		if len(batch) == 1000 {
			file := directoryPrefix + strconv.Itoa(env.Checkpoint.CurrentEpoc) + "/" + time.Now().String() + ".jsonl"
			env.Storage.Put(os.Getenv("BUCKET_NAME"), file, []byte(strings.Join(batch, "\n")))
			batch = make([]string, 0)
		}
	}
}
