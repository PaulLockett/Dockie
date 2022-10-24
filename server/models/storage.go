package models

import (
	"log"

	surrealdb "github.com/surrealdb/surrealdb.go"
)

type StorageModel struct {
	DB *surrealdb.DB
}

type LocalUserStub struct {
	InNextEpoc      bool   `json:"inNextEpoc"`
	TimesUsed       int    `json:"TimesUsed"`
	InServedStorage bool   `json:"InServedStorage"`
	UserAuthKey     string `json:"UserAuthKey"`
}

type LocalCheckpoint struct {
	CurrentEpoc  int                      `json:"CurrentEpoc"`
	NumKeptUsers int                      `json:"NumKeptUsers"`
	UserMap      map[string]LocalUserStub `json:"UserMap"`
}

// Setup connects to the database
func (s *StorageModel) Setup(url string, user string, password string, nameSpace string, database string) error {
	if DB, err := surrealdb.New(url); err != nil {
		log.Println(err)
		return err
	} else {
		log.Println("Connected to database")
		log.Println(DB)
		log.Println("--------------------------------------------------------------------------------------------------")
		if _, err := DB.Signin(map[string]any{"user": user, "pass": password}); err != nil {
			log.Println(err)
			return err
		}
		log.Println("Signed in to database")
		DB.Use(nameSpace, database)
		log.Println("Using database")
		s.DB = DB
	}
	return nil
}

// Put writes a key-value pair to the database.
func (s *StorageModel) Put(name string, data map[string]interface{}) error {
	if _, err := s.DB.Create(name, data); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// GetCheckpoint returns the go map representation of the checkpoint JSON file.
func (s *StorageModel) GetCheckpoint() (LocalCheckpoint, error) {
	storageCheckpoint, err := s.DB.Select("checkpoint:2")
	if err != nil {
		log.Println(err)
		return LocalCheckpoint{}, err
	}
	log.Println(storageCheckpoint.(map[string]interface{})["checkpoint"])

	return parseCheckpoint(storageCheckpoint.(map[string]interface{})["checkpoint"].(LocalCheckpoint))
}

// PutCheckpoint writes the go map representation of the checkpoint JSON file.
func (s *StorageModel) PutCheckpoint(checkpoint LocalCheckpoint) error {
	if _, err := s.DB.Update("checkpoint:2", map[string]interface{}{"checkpoint": checkpoint}); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func parseCheckpoint(checkpoint LocalCheckpoint) (LocalCheckpoint, error) {
	var localCheckpoint LocalCheckpoint
	countKeptUsers := 0
	localCheckpoint.CurrentEpoc = checkpoint.CurrentEpoc
	localCheckpoint.UserMap = make(map[string]LocalUserStub)
	for userID, user := range checkpoint.UserMap {
		localCheckpoint.UserMap[userID] = LocalUserStub{
			InNextEpoc:      user.InNextEpoc,
			TimesUsed:       user.TimesUsed,
			InServedStorage: user.InServedStorage,
			UserAuthKey:     user.UserAuthKey,
		}
		if user.InNextEpoc {
			countKeptUsers++
		}
	}
	localCheckpoint.NumKeptUsers = countKeptUsers
	return localCheckpoint, nil
}
