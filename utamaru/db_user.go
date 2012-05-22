package utamaru

import (
	"time"
	"appengine"
	"appengine/datastore"
)

type TwitterRequestToken struct {
	OauthToken string
	OauthSecret string
	Created_At time.Time
}

func NewRequestToken(m map[string]string) TwitterRequestToken {
	var requestToken TwitterRequestToken
	requestToken.OauthToken = m["oauth_token"]
	requestToken.OauthSecret = m["oauth_secret"]
	return requestToken
}

func SaveRequestToken(c appengine.Context, requestToken TwitterRequestToken) error {
	c.Debugf("SaveRequestToken requestToken: %s", requestToken.OauthToken)
	key := datastore.NewKey(c, "TwitterRequestToken", requestToken.OauthToken, 0, nil)

	requestToken.Created_At = time.Now()

	if _, err := datastore.Put(c, key, &requestToken); err != nil {
		c.Errorf("SaveRequestToken failed to put: %v", err)
		return err
	}

	c.Debugf("SaveRequestToken ok")
	return nil
}

func FindRequestToken(c appengine.Context, requestTokenString string) (TwitterRequestToken, error) {
	c.Debugf("FindRequestToken requestToken: %s", requestTokenString)
	//search
	key := datastore.NewKey(c, "TwitterRequestToken", requestTokenString, 0, nil)

	var requestToken TwitterRequestToken
	if err := datastore.Get(c, key, &requestToken); err != nil {
		c.Errorf("FindRequestToken failed to get: %v", err)
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
	Created_At time.Time
}

func NewTwitterUser(m map[string]string) TwitterUser {
	var user TwitterUser
	user.Id = m["user_id"]
	user.ScreenName = m["screen_name"]
	user.OauthToken = m["oauth_token"]
	user.OauthSecret = m["oauth_token_secret"]
	return user
}

func SaveUser(c appengine.Context, user TwitterUser) error {
	c.Debugf("SaveUser screen_name: %s", user.ScreenName)
	key := datastore.NewKey(c, "TwitterUser", user.SessionId, 0, nil)

	user.Created_At = time.Now()

	if _, err := datastore.Put(c, key, &user); err != nil {
		c.Errorf("SaveUser failed to put: %v", err)
		return err
	}

	c.Debugf("SaveUser ok")
	return nil
}

func FindUser(c appengine.Context, sessionId string) (TwitterUser, error) {
	c.Debugf("FindUser sessionId: %s", sessionId)
	key := datastore.NewKey(c, "TwitterUser", sessionId, 0, nil)

	var user TwitterUser
	if err := datastore.Get(c, key, &user); err != nil {
		c.Errorf("FindUser failed to get: %v", err)
		return user, err
	}

	c.Debugf("FindUser ok")
	return user, nil
}

func DeleteUser(c appengine.Context, sessionId string) error {
	c.Debugf("DeleteUser sessionId: %s", sessionId)
	key := datastore.NewKey(c, "TwitterUser", sessionId, 0, nil)

	if err := datastore.Delete(c, key); err != nil {
		c.Errorf("DeleteUser failed to get: %v", err)
		return err
	}

	c.Debugf("DeleteUser ok")
	return nil
}
