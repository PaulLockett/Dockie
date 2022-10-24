package lib

import (
	"app/models"
	"context"
	"encoding/json"
	"errors"
	"log"

	twitter "github.com/g8rswimmer/go-twitter/v2"
)

func (env *Env) expandFollowers(user UserRequest) (UserRequest, bool) {
	log.Printf("Expanding followers for userID %s", user.userID)
	opts := twitter.UserFollowersLookupOpts{
		UserFields: []twitter.UserField{twitter.UserFieldDescription, twitter.UserFieldEntities, twitter.UserFieldLocation, twitter.UserFieldName, twitter.UserFieldProfileImageURL, twitter.UserFieldURL, twitter.UserFieldCreatedAt, twitter.UserFieldID, twitter.UserFieldPinnedTweetID, twitter.UserFieldProtected, twitter.UserFieldPublicMetrics, twitter.UserFieldUserName, twitter.UserFieldVerified, twitter.UserFieldWithHeld},
		MaxResults: 1000,
	}

	if user.nextToken != "" {
		opts.PaginationToken = user.nextToken
	}

	env.userFollowerRequestLimiter.Take()
	userResponse, err := env.TwitterClient.UserFollowersLookup(context.Background(), user.userID, opts)
	env.reportError(user.userID, err, userResponse.Raw.Errors)

	if rateLimit, has := twitter.RateLimitFromError(err); has && rateLimit.Remaining == 0 {
		return UserRequest{userID: user.userID, nextToken: user.nextToken}, false
	}

	if userResponse.Meta.NextToken != "" {
		return UserRequest{userID: user.userID, nextToken: userResponse.Meta.NextToken}, false
	}

	dictionaries := userResponse.Raw.UserDictionaries()
	followerMappings := make([]FollowMap, len(dictionaries))
	for _, dictionary := range dictionaries {
		followerMappings = append(followerMappings, FollowMap{
			UserID:     "user_data:" + user.userID,
			FollowerID: "user_data:" + dictionary.User.ID,
		})
	}
	env.sendUserData(dictionaries, false, &followerMappings)
	return UserRequest{}, true
}

func (env *Env) expandFriends(user UserRequest) (UserRequest, bool) {
	log.Printf("Expanding friends for userID %s", user.userID)
	opts := twitter.UserFollowingLookupOpts{
		UserFields: []twitter.UserField{twitter.UserFieldDescription, twitter.UserFieldEntities, twitter.UserFieldLocation, twitter.UserFieldName, twitter.UserFieldProfileImageURL, twitter.UserFieldURL, twitter.UserFieldCreatedAt, twitter.UserFieldID, twitter.UserFieldPinnedTweetID, twitter.UserFieldProtected, twitter.UserFieldPublicMetrics, twitter.UserFieldUserName, twitter.UserFieldVerified, twitter.UserFieldWithHeld},
		MaxResults: 1000,
	}

	if user.nextToken != "" {
		opts.PaginationToken = user.nextToken
	}

	env.userFriendRequestLimiter.Take()
	userResponse, err := env.TwitterClient.UserFollowingLookup(context.Background(), user.userID, opts)
	env.reportError(user.userID, err, userResponse.Raw.Errors)

	if rateLimit, has := twitter.RateLimitFromError(err); has && rateLimit.Remaining == 0 {
		return UserRequest{userID: user.userID, nextToken: user.nextToken}, false
	}

	if userResponse.Meta.NextToken != "" {
		return UserRequest{userID: user.userID, nextToken: userResponse.Meta.NextToken}, false
	}

	dictionaries := userResponse.Raw.UserDictionaries()
	followerMappings := make([]FollowMap, len(dictionaries))
	for _, dictionary := range dictionaries {
		followerMappings = append(followerMappings, FollowMap{
			UserID:     "user_data:" + dictionary.User.ID,
			FollowerID: "user_data:" + user.userID,
		})
	}
	env.sendUserData(dictionaries, false, &followerMappings)

	return UserRequest{}, true

}

func (env *Env) expandUsers(userIDs []string) bool {
	log.Printf("Expanding users for %d users", len(userIDs))
	opts := twitter.UserLookupOpts{
		UserFields: []twitter.UserField{twitter.UserFieldDescription, twitter.UserFieldEntities, twitter.UserFieldLocation, twitter.UserFieldName, twitter.UserFieldProfileImageURL, twitter.UserFieldURL, twitter.UserFieldCreatedAt, twitter.UserFieldID, twitter.UserFieldPinnedTweetID, twitter.UserFieldProtected, twitter.UserFieldPublicMetrics, twitter.UserFieldUserName, twitter.UserFieldVerified, twitter.UserFieldWithHeld},
	}

	env.batchUserRequestLimiter.Take()
	userResponse, err := env.TwitterClient.UserLookup(context.Background(), userIDs, opts)
	env.reportError("(batch user lookup)", err, userResponse.Raw.Errors)

	if rateLimit, has := twitter.RateLimitFromError(err); has && rateLimit.Remaining == 0 {
		env.batchUserRequestChan <- batchUserRequest{userIDs: userIDs}
		return false
	}

	dictionaries := userResponse.Raw.UserDictionaries()
	env.sendUserData(dictionaries, true, &[]FollowMap{})

	return true
}

func (env *Env) reportError(userId string, err error, partialErrors []*twitter.ErrorObj) {
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

func (env *Env) sendUserData(dictionaries map[string]*twitter.UserDictionary, keepFresh bool, mappings *[]FollowMap) {
	if len(dictionaries) != 0 {
		log.Printf("Sending %d user dictionaries", len(dictionaries))
		for _, dictionary := range dictionaries {
			userData := dictionary.User
			env.Checkpoint.UserMap[dictionary.User.ID] = models.LocalUserStub{
				InNextEpoc:      keepFresh,
				TimesUsed:       0,
				InServedStorage: true,
				UserAuthKey:     "",
			}
			env.dataChan <- userData
		}
		env.Storage.PutCheckpoint(env.Checkpoint)
	}
	if len(*mappings) != 0 {
		log.Printf("Sending %d follower mappings", len(*mappings))
		for _, mapping := range *mappings {
			env.dataChan <- mapping
		}
	}
}
