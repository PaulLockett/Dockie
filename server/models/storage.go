package models

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"

	"cloud.google.com/go/storage"
)

type StorageModel struct {
	Client *storage.Client
	ctx    context.Context
}

type StorageCheckpoint struct {
	CurrentEpoc int               `json:"CurrentEpoc"`
	UserList    []StorageUserStub `json:"UserList"`
}

type StorageUserStub struct {
	UserID          string `json:"userID"`
	InNextEpoc      bool   `json:"inNextEpoc"`
	TimesUsed       string `json:"TimesUsed"`
	InServedStorage bool   `json:"InServedStorage"`
	UserAuthKey     string `json:"UserAuthKey"`
}

type LocalUserStub struct {
	InNextEpoc      bool   `json:"inNextEpoc"`
	TimesUsed       int    `json:"TimesUsed"`
	InServedStorage bool   `json:"InServedStorage"`
	UserAuthKey     string `json:"UserAuthKey"`
}

type LocalCheckpoint struct {
	CurrentEpoc int                      `json:"CurrentEpoc"`
	UserMap     map[string]LocalUserStub `json:"UserMap"`
}

func (g *StorageModel) Setup(ctx context.Context) error {
	var err error
	if g.Client, err = storage.NewClient(ctx); err != nil {
		return err
	}
	g.ctx = ctx
	return nil
}

func (g *StorageModel) get(bucket, object string) ([]byte, error) {
	rc, err := g.Client.Bucket(bucket).Object(object).NewReader(g.ctx)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return ioutil.ReadAll(rc)
}

func (g *StorageModel) Put(bucket, object string, data []byte) error {
	wc := g.Client.Bucket(bucket).Object(object).NewWriter(g.ctx)
	defer wc.Close()
	_, err := wc.Write(data)
	return err
}

// GetCheckpoint returns the go map representation of the checkpoint JSON file.
func (g *StorageModel) GetCheckpoint() (LocalCheckpoint, error) {
	data, err := g.get(os.Getenv("BUCKET_NAME"), "checkpoint.json")
	if err != nil {
		return LocalCheckpoint{}, err
	}
	return parseCheckpoint(data)
}

// PutCheckpoint writes the go map representation of the checkpoint JSON file.
func (g *StorageModel) PutCheckpoint(checkpoint LocalCheckpoint) error {
	data, err := formatCheckpoint(checkpoint)
	if err != nil {
		return err
	}
	return g.Put(os.Getenv("BUCKET_NAME"), "checkpoint.json", data)
}

func parseCheckpoint(data []byte) (LocalCheckpoint, error) {
	var checkpoint StorageCheckpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return LocalCheckpoint{}, err
	}

	var localCheckpoint LocalCheckpoint
	localCheckpoint.CurrentEpoc = checkpoint.CurrentEpoc
	localCheckpoint.UserMap = make(map[string]LocalUserStub)
	for _, user := range checkpoint.UserList {
		timesUsed, err := strconv.Atoi(user.TimesUsed)
		if err != nil {
			return LocalCheckpoint{}, err
		}
		localCheckpoint.UserMap[user.UserID] = LocalUserStub{
			InNextEpoc:      user.InNextEpoc,
			TimesUsed:       timesUsed,
			InServedStorage: user.InServedStorage,
			UserAuthKey:     user.UserAuthKey,
		}
	}
	return localCheckpoint, nil
}

func formatCheckpoint(checkpoint LocalCheckpoint) ([]byte, error) {
	var storageCheckpoint StorageCheckpoint
	storageCheckpoint.CurrentEpoc = checkpoint.CurrentEpoc
	storageCheckpoint.UserList = make([]StorageUserStub, len(checkpoint.UserMap))
	i := 0
	for userID, user := range checkpoint.UserMap {
		storageCheckpoint.UserList[i] = struct {
			UserID          string `json:"userID"`
			InNextEpoc      bool   `json:"inNextEpoc"`
			TimesUsed       string `json:"TimesUsed"`
			InServedStorage bool   `json:"InServedStorage"`
			UserAuthKey     string `json:"UserAuthKey"`
		}{
			UserID:          userID,
			InNextEpoc:      user.InNextEpoc,
			TimesUsed:       strconv.Itoa(user.TimesUsed),
			InServedStorage: user.InServedStorage,
			UserAuthKey:     user.UserAuthKey,
		}
		i++
	}
	return json.Marshal(storageCheckpoint)
}
