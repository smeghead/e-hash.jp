package utamaru

import (
	"os"
	"appengine"
	"appengine/datastore"
)

type TwitterConf struct {
	ConsumerKey string
	ConsumerSecret string
	AccessToken string
	AccessTokenSecret string
}

func GetTwitterConf(c appengine.Context) (TwitterConf, os.Error) {
	c.Infof("GetTwitterConf")
	conf := new(TwitterConf)
	key := datastore.NewKey("TwitterConf", "singleton", 0, nil)

	if err := datastore.Get(c, key, conf); err != nil {
		c.Infof("GetTwitterConf failed to load: %v", err)
		c.Infof("GetTwitterConf try to initialize config.")
		if _, err = datastore.Put(c, key, conf); err != nil {
			c.Errorf("GetTwitterConf failed to load: %v", err)
		}
		return *conf, err
	}
	return *conf, nil
}