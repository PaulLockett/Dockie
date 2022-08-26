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
	userFollowerChan     chan UserRequest
	userFriendChan       chan UserRequest
	userDataChan         chan PushData
	batchUserRequestChan chan batchUserRequest
	followMapChan        chan PushData
	pushDataCloseChecks  []bool
	Storage              interface {
		Setup(ctx context.Context, ErrorLogger *log.Logger) error
		GetCheckpoint() (models.LocalCheckpoint, error)
		PutCheckpoint(checkpoint models.LocalCheckpoint) error
		Put(bucket, object string, data []byte) error
	}
}

type UserRequest struct {
	userID         string
	nextToken      string
	rateLimitReset time.Time
	closingUser    bool
	shouldClose    bool
}

type batchUserRequest struct {
	userIDs        []string
	rateLimitReset time.Time
}

type FollowMap struct {
	UserID     string
	FollowerID string
}

type PushData struct {
	data        string
	shouldClose bool
}

func (env *Env) FollowerExpander(userFollowerRequestChan <-chan UserRequest) {
	env.RunLogger.Println("in FollowerExpander")
	defer env.wg.Done()

	for userRequest := range userFollowerRequestChan {
		if !userRequest.rateLimitReset.IsZero() {
			env.RunLogger.Printf("sleeping for %s in FollowerExpander", time.Until(userRequest.rateLimitReset))
			time.Sleep(time.Until(userRequest.rateLimitReset))
			env.RunLogger.Printf("done sleeping in FollowerExpander")
		}
		env.expandFollowers(userRequest)
		if userRequest.shouldClose {
			close(env.userFollowerChan)
			return
		}
	}
}

func (env *Env) FriendExpander(userFriendRequestChan <-chan UserRequest) {
	env.RunLogger.Println("in FriendExpander")
	defer env.wg.Done()

	for userRequest := range userFriendRequestChan {
		if !userRequest.rateLimitReset.IsZero() {
			env.RunLogger.Printf("sleeping for %s in FriendExpander", time.Until(userRequest.rateLimitReset))
			time.Sleep(time.Until(userRequest.rateLimitReset))
			env.RunLogger.Printf("done sleeping in FriendExpander")
		}
		env.expandFriends(userRequest)
		if userRequest.shouldClose {
			close(env.userFriendChan)
			return
		}
	}
}

func (env *Env) BatchUserExpander(batchUserRequestChan <-chan batchUserRequest) {
	env.RunLogger.Println("in BatchUserExpander")
	defer env.wg.Done()

	for batchUserRequest := range batchUserRequestChan {
		if !batchUserRequest.rateLimitReset.IsZero() {
			env.RunLogger.Printf("sleeping for %s in BatchUserExpander", time.Until(batchUserRequest.rateLimitReset))
			time.Sleep(time.Until(batchUserRequest.rateLimitReset))
			env.RunLogger.Printf("done sleeping in BatchUserExpander")
		}
		env.expandUsers(batchUserRequest.userIDs)
	}
}

func (env *Env) UserDataSaver(userDataChan <-chan PushData) {
	env.RunLogger.Println("in UserDataSaver")
	defer env.wg.Done()

	env.saveUserData(userDataChan, "user_data/")
	close(env.userDataChan)

	env.RunLogger.Println("done in UserDataSaver")
}

func (env *Env) FollowMappingSaver(followMapChan <-chan PushData) {
	env.RunLogger.Println("in FollowMappingSaver")
	defer env.wg.Done()

	env.saveUserData(followMapChan, "follow_map/")
	close(env.followMapChan)

	env.RunLogger.Println("done in FollowMappingSaver")
}

func (env *Env) saveUserData(Chan <-chan PushData, filePrefix string) {
	batch := make([]string, 0)
	for obj := range Chan {
		batch = append(batch, obj.data)
		if len(batch) == 1000 {
			file := filePrefix + strconv.Itoa(env.Checkpoint.CurrentEpoc) + "/" + strconv.Itoa(int(time.Now().Unix())) + ".jsonl"
			env.Storage.Put(os.Getenv("BUCKET_NAME"), file, []byte(strings.Join(batch, "\n")))
			batch = make([]string, 0)
		}
		if obj.shouldClose {
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
	// setup env channels
	// 360 is max number of Twitter API calls per 15 minutes per 24 hour period per Authenticated User
	env.userFollowerChan = make(chan UserRequest, 360)
	env.userFriendChan = make(chan UserRequest, 360)
	env.batchUserRequestChan = make(chan batchUserRequest, 360)

	env.userDataChan = make(chan PushData, 100000)
	env.followMapChan = make(chan PushData, 100000)

	env.pushDataCloseChecks = make([]bool, 2)

	// setup env waitgroups
	env.wg = &sync.WaitGroup{}
	env.wg.Add(5)

	// start expander workers
	go env.FollowerExpander(env.userFollowerChan)
	go env.FriendExpander(env.userFriendChan)
	go env.BatchUserExpander(env.batchUserRequestChan)

	go env.UserDataSaver(env.userDataChan)
	go env.FollowMappingSaver(env.followMapChan)

	batch := make([]string, 0)
	countKeptUsersSeen := 0
	for userID, userStub := range env.Checkpoint.UserMap {
		if userStub.InNextEpoc {
			countKeptUsersSeen++
			closingUser := false
			if countKeptUsersSeen == env.Checkpoint.NumKeptUsers {
				closingUser = true
			}
			batch = append(batch, userID)
			if len(batch) == 1000 {
				env.batchUserRequestChan <- batchUserRequest{
					userIDs:        batch,
					rateLimitReset: time.Time{},
				}
				batch = make([]string, 0)
			}
			env.userFollowerChan <- UserRequest{
				userID:         userID,
				nextToken:      "",
				rateLimitReset: time.Time{},
				closingUser:    closingUser,
				shouldClose:    false,
			}
			env.userFriendChan <- UserRequest{
				userID:         userID,
				nextToken:      "",
				rateLimitReset: time.Time{},
				closingUser:    closingUser,
				shouldClose:    false,
			}
		}
	}
	if len(batch) > 0 {
		env.batchUserRequestChan <- batchUserRequest{
			userIDs:        batch,
			rateLimitReset: time.Time{},
		}
	}
	close(env.batchUserRequestChan)

	env.wg.Wait()
	env.RunLogger.Println("done Refresh")
}
