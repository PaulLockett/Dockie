package lib

import (
	"app/models"
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	twitter "github.com/g8rswimmer/go-twitter/v2"
)

func (env *Env) expandFollowers(userId string, nextToken string) {

	opts := twitter.UserFollowersLookupOpts{
		UserFields: []twitter.UserField{twitter.UserFieldDescription, twitter.UserFieldEntities, twitter.UserFieldLocation, twitter.UserFieldName, twitter.UserFieldProfileImageURL, twitter.UserFieldURL, twitter.UserFieldCreatedAt, twitter.UserFieldID, twitter.UserFieldPinnedTweetID, twitter.UserFieldProtected, twitter.UserFieldPublicMetrics, twitter.UserFieldUserName, twitter.UserFieldVerified, twitter.UserFieldWithHeld},
		MaxResults: 1000,
	}

	if nextToken != "" {
		opts.PaginationToken = nextToken
	}

	userResponse, err := env.TwitterClient.UserFollowersLookup(context.Background(), userId, opts)
	reportError(userId, err, userResponse.Raw.Errors)

	if rateLimit, has := twitter.RateLimitFromError(err); has && rateLimit.Remaining == 0 {
		env.userFollowerChan <- userRequest{userID: userId, nextToken: nextToken, rateLimitReset: rateLimit.Reset.Time()}
		return
	}

	if userResponse.Meta.NextToken != "" {
		env.userFollowerChan <- userRequest{userID: userId, nextToken: userResponse.Meta.NextToken, rateLimitReset: time.Time{}}
	}

	dictionaries := userResponse.Raw.UserDictionaries()
	followerMappings := make([]FollowMap, len(dictionaries))
	for _, dictionary := range dictionaries {
		followerMappings = append(followerMappings, FollowMap{
			UserID:     userId,
			FollowerID: dictionary.User.ID,
		})
	}
	env.sendUserData(dictionaries, false, followerMappings)

}

func (env *Env) expandFriends(userId string, nextToken string) {

	opts := twitter.UserFollowingLookupOpts{
		UserFields: []twitter.UserField{twitter.UserFieldDescription, twitter.UserFieldEntities, twitter.UserFieldLocation, twitter.UserFieldName, twitter.UserFieldProfileImageURL, twitter.UserFieldURL, twitter.UserFieldCreatedAt, twitter.UserFieldID, twitter.UserFieldPinnedTweetID, twitter.UserFieldProtected, twitter.UserFieldPublicMetrics, twitter.UserFieldUserName, twitter.UserFieldVerified, twitter.UserFieldWithHeld},
		MaxResults: 1000,
	}

	if nextToken != "" {
		opts.PaginationToken = nextToken
	}

	userResponse, err := env.TwitterClient.UserFollowingLookup(context.Background(), userId, opts)
	reportError(userId, err, userResponse.Raw.Errors)

	if rateLimit, has := twitter.RateLimitFromError(err); has && rateLimit.Remaining == 0 {
		env.userFriendChan <- userRequest{userID: userId, nextToken: nextToken, rateLimitReset: rateLimit.Reset.Time()}
		return
	}

	if userResponse.Meta.NextToken != "" {
		env.userFriendChan <- userRequest{userID: userId, nextToken: userResponse.Meta.NextToken, rateLimitReset: time.Time{}}
	}
	dictionaries := userResponse.Raw.UserDictionaries()
	followerMappings := make([]FollowMap, len(dictionaries))
	for _, dictionary := range dictionaries {
		followerMappings = append(followerMappings, FollowMap{
			UserID:     dictionary.User.ID,
			FollowerID: userId,
		})
	}
	env.sendUserData(dictionaries, false, followerMappings)

}

func (env *Env) expandUsers(userIDs []string) {

	opts := twitter.UserLookupOpts{
		UserFields: []twitter.UserField{twitter.UserFieldDescription, twitter.UserFieldEntities, twitter.UserFieldLocation, twitter.UserFieldName, twitter.UserFieldProfileImageURL, twitter.UserFieldURL, twitter.UserFieldCreatedAt, twitter.UserFieldID, twitter.UserFieldPinnedTweetID, twitter.UserFieldProtected, twitter.UserFieldPublicMetrics, twitter.UserFieldUserName, twitter.UserFieldVerified, twitter.UserFieldWithHeld},
	}

	userResponse, err := env.TwitterClient.UserLookup(context.Background(), userIDs, opts)
	reportError("(batch user lookup)", err, userResponse.Raw.Errors)

	if rateLimit, has := twitter.RateLimitFromError(err); has && rateLimit.Remaining == 0 {
		env.batchUserRequestChan <- batchUserRequest{userIDs: userIDs, rateLimitReset: rateLimit.Reset.Time()}
		return
	}

	dictionaries := userResponse.Raw.UserDictionaries()
	env.sendUserData(dictionaries, true, []FollowMap{})
}

func reportError(userId string, err error, partialErrors []*twitter.ErrorObj) {
	jsonErr := &json.UnsupportedValueError{}
	ResponseDecodeError := &twitter.ResponseDecodeError{}

	switch {
	case errors.Is(err, twitter.ErrParameter):
		// handle a parameter error
		log.Printf("Parameter error in expandFollowers for userID %s : %s", userId, err)
	case errors.As(err, &jsonErr):
		// handle a json error
		log.Printf("JSON error in expandFollowers for userID %s : %s", userId, err)
	case errors.As(err, &ResponseDecodeError):
		// handle response decode error
		log.Printf("Response decode error in expandFollowers for userID %s : %s", userId, err)
	case errors.As(err, &twitter.HTTPError{}):
		// handle http response error
		log.Printf("HTTP error in expandFollowers for userID %s : %s", userId, err)
	case err != nil:
		// handle other errors
		log.Printf("Other error in expandFollowers for userID %s : %s", userId, err)
	default:
		// happy path
	}

	if len(partialErrors) > 0 {
		// handle partial errors
		for _, err := range partialErrors {
			log.Printf("Partial error in expandFollowers for userID %s : %s", userId, err)
		}
	}
}

func (env *Env) sendUserData(dictionaries map[string]*twitter.UserDictionary, keepFresh bool, mappings []FollowMap) {
	for _, dictionary := range dictionaries {
		userData, err := json.Marshal(dictionary)
		env.Checkpoint.UserMap[dictionary.User.ID] = models.LocalUserStub{
			InNextEpoc:      keepFresh,
			TimesUsed:       0,
			InServedStorage: true,
			UserAuthKey:     "",
		}
		if err != nil {
			log.Printf("Error marshalling user data : %s", err)
		}
		env.userDataChan <- string(userData)
	}
	for _, mapping := range mappings {
		mappingJSON, err := json.Marshal(mapping)
		if err != nil {
			log.Printf("Error marshalling mapping data : %s", err)
		}
		env.followMapChan <- string(mappingJSON)
	}
}
