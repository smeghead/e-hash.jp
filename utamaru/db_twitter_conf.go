package utamaru

import (
	"appengine"
	"appengine/datastore"
)

type TwitterConf struct {
	Url string
	ConsumerKey string
	ConsumerSecret string
	AccessToken string
	AccessTokenSecret string
	GoogleShortUrlApiKey string
}

func GetTwitterConf(c appengine.Context) (TwitterConf, error) {
	c.Infof("GetTwitterConf")
	conf := new(TwitterConf)
	key := datastore.NewKey(c, "TwitterConf", "singleton", 0, nil)

	if err := datastore.Get(c, key, conf); err != nil {
		c.Infof("GetTwitterConf failed to load: %v", err)
		c.Infof("GetTwitterConf try to initialize config.")
		if _, err = datastore.Put(c, key, conf); err != nil {
			c.Errorf("GetTwitterConf failed to load: %v", err)
		}
		return *conf, err
	}
	if _, err := datastore.Put(c, key, conf); err != nil {
		c.Errorf("GetTwitterConf failed to load: %v", err)
	}
	return *conf, nil
}

