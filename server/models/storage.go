package models

import (
	"context"
	"encoding/json"
	"io/ioutil"

	"cloud.google.com/go/storage"
)

type StorageModel struct {
	Client *storage.Client
	ctx    context.Context
}

type StorageCheckpoint struct {
	CurrentEpoc int               `json:"CurrentEpoc,omitempty"`
	UserList    []StorageUserStub `json:"UserList,omitempty"`
}

type StorageUserStub struct {
	UserID          string `json:"userID,omitempty"`
	InNextEpoc      bool   `json:"inNextEpoc,omitempty"`
	TimesUsed       string `json:"TimesUsed,omitempty"`
	InServedStorage bool   `json:"InServedStorage,omitempty"`
	UserAuthKey     string `json:"UserAuthKey,omitempty"`
}

type LocalUserStub struct {
	InNextEpoc      bool   `json:"inNextEpoc,omitempty"`
	TimesUsed       string `json:"TimesUsed,omitempty"`
	InServedStorage bool   `json:"InServedStorage,omitempty"`
	UserAuthKey     string `json:"UserAuthKey,omitempty"`
}

type LocalCheckpoint struct {
	CurrentEpoc int                      `json:"CurrentEpoc,omitempty"`
	UserMap     map[string]LocalUserStub `json:"UserMap,omitempty"`
}

func (g *StorageModel) Setup(ctx context.Context) error {
	var err error
	if g.Client, err = storage.NewClient(ctx); err != nil {
		return err
	}
	g.ctx = ctx
	return nil
}

func (g *StorageModel) get(ctx context.Context, bucket, object string) ([]byte, error) {
	rc, err := g.Client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return ioutil.ReadAll(rc)
}

func (g *StorageModel) put(ctx context.Context, bucket, object string, data []byte) error {
	wc := g.Client.Bucket(bucket).Object(object).NewWriter(ctx)
	defer wc.Close()
	_, err := wc.Write(data)
	return err
}

// GetCheckpoint returns the go map representation of the checkpoint JSON file.
func (g *StorageModel) GetCheckpoint() (LocalCheckpoint, error) {
	data, err := g.get(g.ctx, "twitter_users_v1", "checkpoint.json")
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
	return g.put(g.ctx, "twitter_users_v1", "checkpoint.json", data)
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
		localCheckpoint.UserMap[user.UserID] = LocalUserStub{
			InNextEpoc:      user.InNextEpoc,
			TimesUsed:       user.TimesUsed,
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
			UserID          string `json:"userID,omitempty"`
			InNextEpoc      bool   `json:"inNextEpoc,omitempty"`
			TimesUsed       string `json:"TimesUsed,omitempty"`
			InServedStorage bool   `json:"InServedStorage,omitempty"`
			UserAuthKey     string `json:"UserAuthKey,omitempty"`
		}{
			UserID:          userID,
			InNextEpoc:      user.InNextEpoc,
			TimesUsed:       user.TimesUsed,
			InServedStorage: user.InServedStorage,
			UserAuthKey:     user.UserAuthKey,
		}
		i++
	}
	return json.Marshal(storageCheckpoint)
}
