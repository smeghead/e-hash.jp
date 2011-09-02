package utamaru

import (
	"os"
	"time"
	"appengine"
	"appengine/datastore"
)

type TwitterRequestToken struct {
	OauthToken string
	OauthSecret string
	Created_At datastore.Time
}

func NewRequestToken(m map[string]string) TwitterRequestToken {
	var requestToken TwitterRequestToken
	requestToken.OauthToken = m["oauth_token"]
	requestToken.OauthSecret = m["oauth_secret"]
	return requestToken
}

func SaveRequestToken(c appengine.Context, requestToken TwitterRequestToken) os.Error {
	c.Debugf("SaveRequestToken requestToken: %s", requestToken.OauthToken)
	key := datastore.NewKey("TwitterRequestToken", requestToken.OauthToken, 0, nil)

	requestToken.Created_At = datastore.SecondsToTime(time.Seconds())

	if _, err := datastore.Put(c, key, &requestToken); err != nil {
		c.Errorf("SaveRequestToken failed to put: %v", err.String())
		return err
	}

	c.Debugf("SaveRequestToken ok")
	return nil
}

func FindRequestToken(c appengine.Context, requestTokenString string) (TwitterRequestToken, os.Error) {
	c.Debugf("FindRequestToken requestToken: %s", requestTokenString)
	//search
	key := datastore.NewKey("TwitterRequestToken", requestTokenString, 0, nil)

	var requestToken TwitterRequestToken
	if err := datastore.Get(c, key, &requestToken); err != nil {
		c.Errorf("FindRequestToken failed to get: %v", err.String())
		return requestToken, err
	}

	c.Debugf("FindRequestToken ok")
	return requestToken, nil
}

type TwitterUser struct {
	SessionId string
	Id string
	ScreenName string
	OauthToken string
	OauthSecret string
	Created_At datastore.Time
}

func NewTwitterUser(m map[string]string) TwitterUser {
	var user TwitterUser
	user.Id = m["user_id"]
	user.ScreenName = m["screen_name"]
	user.OauthToken = m["oauth_token"]
	user.OauthSecret = m["oauth_token_secret"]
	return user
}

func SaveUser(c appengine.Context, user TwitterUser) os.Error {
	c.Debugf("SaveUser screen_name: %s", user.ScreenName)
	key := datastore.NewKey("TwitterUser", user.SessionId, 0, nil)

	user.Created_At = datastore.SecondsToTime(time.Seconds())

	if _, err := datastore.Put(c, key, &user); err != nil {
		c.Errorf("SaveUser failed to put: %v", err.String())
		return err
	}

	c.Debugf("SaveUser ok")
	return nil
}

func FindUser(c appengine.Context, sessionId string) (TwitterUser, os.Error) {
	c.Debugf("FindUser sessionId: %s", sessionId)
	key := datastore.NewKey("TwitterUser", sessionId, 0, nil)

	var user TwitterUser
	if err := datastore.Get(c, key, &user); err != nil {
		c.Errorf("FindUser failed to get: %v", err.String())
		return user, err
	}

	c.Debugf("FindUser ok")
	return user, nil
}
